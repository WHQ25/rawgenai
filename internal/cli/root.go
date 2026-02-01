package cli

import (
	"fmt"

	"github.com/WHQ25/rawgenai/internal/cli/elevenlabs"
	"github.com/WHQ25/rawgenai/internal/cli/google"
	"github.com/WHQ25/rawgenai/internal/cli/grok"
	"github.com/WHQ25/rawgenai/internal/cli/openai"
	"github.com/spf13/cobra"
)

// Version info set by goreleaser ldflags
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:     "rawgenai",
	Short:   "CLI tool for AI agents to access raw AI capabilities",
	Long:    "A CLI tool designed for AI agents to access raw AI capabilities including TTS, STT, Image and Video generation.",
	Version: version,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("rawgenai %s\n", version)
		fmt.Printf("  commit: %s\n", commit)
		fmt.Printf("  built:  %s\n", date)
	},
}

func init() {
	rootCmd.AddCommand(openai.Cmd)
	rootCmd.AddCommand(google.Cmd)
	rootCmd.AddCommand(elevenlabs.Cmd)
	rootCmd.AddCommand(grok.Cmd)
	rootCmd.AddCommand(versionCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
