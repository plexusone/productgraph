// Package config provides configuration loading for ProductGraph services.
package config

import (
	"os"
	"strconv"
)

// Config holds the application configuration.
type Config struct {
	Port      int
	Debug     bool
	Analytics AnalyticsConfig
}

// AnalyticsConfig configures analytics provider integration.
type AnalyticsConfig struct {
	Enabled   bool
	Amplitude AmplitudeConfig
	Mixpanel  MixpanelConfig
}

// AmplitudeConfig configures the Amplitude provider.
type AmplitudeConfig struct {
	Enabled bool
	APIKey  string
}

// MixpanelConfig configures the Mixpanel provider.
type MixpanelConfig struct {
	Enabled bool
	Token   string
}

// Load loads configuration from environment variables.
func Load() *Config {
	return &Config{
		Port:  getEnvInt("PORT", 8080),
		Debug: getEnvBool("DEBUG", false),
		Analytics: AnalyticsConfig{
			Enabled: getEnvBool("ANALYTICS_ENABLED", false),
			Amplitude: AmplitudeConfig{
				Enabled: getEnvBool("AMPLITUDE_ENABLED", false),
				APIKey:  os.Getenv("AMPLITUDE_API_KEY"),
			},
			Mixpanel: MixpanelConfig{
				Enabled: getEnvBool("MIXPANEL_ENABLED", false),
				Token:   os.Getenv("MIXPANEL_TOKEN"),
			},
		},
	}
}

// HasAnalytics returns true if any analytics provider is enabled.
func (c *Config) HasAnalytics() bool {
	if !c.Analytics.Enabled {
		return false
	}
	return c.Analytics.Amplitude.Enabled || c.Analytics.Mixpanel.Enabled
}

func getEnvInt(key string, defaultValue int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultValue
}

//nolint:unparam // defaultValue is intentionally generic for future use
func getEnvBool(key string, defaultValue bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return defaultValue
}
