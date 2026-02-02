package cli

import (
	"github.com/WHQ25/rawgenai/internal/cli/config"
	"github.com/WHQ25/rawgenai/internal/cli/elevenlabs"
	"github.com/WHQ25/rawgenai/internal/cli/google"
	"github.com/WHQ25/rawgenai/internal/cli/grok"
	"github.com/WHQ25/rawgenai/internal/cli/kling"
	"github.com/WHQ25/rawgenai/internal/cli/luma"
	"github.com/WHQ25/rawgenai/internal/cli/minimax"
	"github.com/WHQ25/rawgenai/internal/cli/openai"
	"github.com/WHQ25/rawgenai/internal/cli/runway"
	"github.com/WHQ25/rawgenai/internal/cli/seed"
	"github.com/spf13/cobra"
)

// Version info set by goreleaser ldflags
var version = "dev"

var rootCmd = &cobra.Command{
	Use:     "rawgenai",
	Short:   "CLI tool for AI agents to access raw AI capabilities",
	Long:    "A CLI tool designed for AI agents to access raw AI capabilities including TTS, STT, Image and Video generation.",
	Version: version,
}

func init() {
	rootCmd.AddCommand(openai.Cmd)
	rootCmd.AddCommand(google.Cmd)
	rootCmd.AddCommand(elevenlabs.Cmd)
	rootCmd.AddCommand(grok.Cmd)
	rootCmd.AddCommand(seed.Cmd)
	rootCmd.AddCommand(kling.Cmd)
	rootCmd.AddCommand(runway.Cmd)
	rootCmd.AddCommand(luma.Cmd)
	rootCmd.AddCommand(minimax.Cmd)
	rootCmd.AddCommand(config.Cmd)
}

func Execute() error {
	return rootCmd.Execute()
}
