package cli

import (
	"github.com/WHQ25/rawgenai/internal/cli/openai"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "rawgenai",
	Short: "CLI tool for AI agents to access raw AI capabilities",
	Long:  "A CLI tool designed for AI agents to access raw AI capabilities including TTS, STT, Image and Video generation.",
}

func init() {
	rootCmd.AddCommand(openai.Cmd)
}

func Execute() error {
	return rootCmd.Execute()
}
