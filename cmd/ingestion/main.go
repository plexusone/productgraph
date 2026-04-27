// Package main is the entry point for the ProductGraph event ingestion service.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/plexusone/omnidxi"
	"github.com/plexusone/productgraph/internal/analytics"
	"github.com/plexusone/productgraph/internal/config"
	"github.com/plexusone/productgraph/internal/events"
)

func main() {
	// Parse flags (override env vars)
	portFlag := flag.Int("port", 0, "HTTP server port (overrides PORT env)")
	debugFlag := flag.Bool("debug", false, "Enable debug logging (overrides DEBUG env)")
	flag.Parse()

	// Load config from environment
	cfg := config.Load()

	// Apply flag overrides
	if *portFlag != 0 {
		cfg.Port = *portFlag
	}
	if *debugFlag {
		cfg.Debug = true
	}

	// Setup logger
	logLevel := slog.LevelInfo
	if cfg.Debug {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)

	// Create publishers
	memoryPub := events.NewMemoryPublisher(logger)
	publisher := createPublisher(logger, cfg, memoryPub)

	// Create handler
	handler := events.NewHandler(logger, publisher)

	// Setup routes
	mux := http.NewServeMux()
	mux.Handle("POST /v1/events", handler)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ready"}`))
	})

	// Add CORS middleware for development
	corsHandler := corsMiddleware(mux)

	// Create server
	addr := fmt.Sprintf(":%d", cfg.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      corsHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("starting ingestion service", "addr", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("shutdown error", "error", err)
	}

	logger.Info("server stopped")
}

// createPublisher creates the event publisher based on configuration.
// If analytics is enabled, returns a MultiPublisher that fans out to both
// the memory publisher and analytics providers.
func createPublisher(logger *slog.Logger, cfg *config.Config, memoryPub events.Publisher) events.Publisher {
	if !cfg.HasAnalytics() {
		logger.Info("analytics disabled, using memory publisher only")
		return memoryPub
	}

	// Build analytics tracker
	tracker := createTracker(logger, cfg)
	if tracker == nil {
		logger.Warn("no analytics providers configured, using memory publisher only")
		return memoryPub
	}

	// Create analytics adapter
	analyticsAdapter := analytics.NewAdapter(tracker)

	// Create multi-publisher for fan-out
	multiPub := events.NewMultiPublisher(memoryPub, analyticsAdapter)
	logger.Info("analytics enabled", "publishers", multiPub.Len())

	return multiPub
}

// createTracker creates the omnidxi tracker based on configured providers.
func createTracker(logger *slog.Logger, cfg *config.Config) omnidxi.Tracker {
	var trackers []omnidxi.Tracker

	// Add Amplitude if configured
	if cfg.Analytics.Amplitude.Enabled && cfg.Analytics.Amplitude.APIKey != "" {
		amp := omnidxi.NewAmplitudeTracker(omnidxi.WithAPIKey(cfg.Analytics.Amplitude.APIKey))
		trackers = append(trackers, amp)
		logger.Info("amplitude provider enabled")
	} else if cfg.Analytics.Amplitude.Enabled {
		logger.Warn("amplitude enabled but API key not set")
	}

	// Add Mixpanel if configured
	if cfg.Analytics.Mixpanel.Enabled && cfg.Analytics.Mixpanel.Token != "" {
		mp := omnidxi.NewMixpanelTracker(omnidxi.WithAPIKey(cfg.Analytics.Mixpanel.Token))
		trackers = append(trackers, mp)
		logger.Info("mixpanel provider enabled")
	} else if cfg.Analytics.Mixpanel.Enabled {
		logger.Warn("mixpanel enabled but token not set")
	}

	if len(trackers) == 0 {
		return nil
	}

	if len(trackers) == 1 {
		return trackers[0]
	}

	return omnidxi.NewMultiTracker(trackers...)
}

// corsMiddleware adds CORS headers for development.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow requests from any origin in development
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}

		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-PG-API-Key")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
