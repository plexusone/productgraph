package cmd

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		version := "dev"
		commit := "unknown"
		date := "unknown"

		if info, ok := debug.ReadBuildInfo(); ok {
			if info.Main.Version != "" && info.Main.Version != "(devel)" {
				version = info.Main.Version
			}
			for _, setting := range info.Settings {
				switch setting.Key {
				case "vcs.revision":
					if len(setting.Value) >= 7 {
						commit = setting.Value[:7]
					}
				case "vcs.time":
					date = setting.Value
				}
			}
		}

		fmt.Fprintf(cmd.OutOrStdout(), "catalogctl %s\n", version)
		fmt.Fprintf(cmd.OutOrStdout(), "  commit: %s\n", commit)
		fmt.Fprintf(cmd.OutOrStdout(), "  built:  %s\n", date)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
