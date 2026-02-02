package kling

import (
	"github.com/WHQ25/rawgenai/internal/cli/kling/video"
	"github.com/WHQ25/rawgenai/internal/cli/kling/voice"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "kling",
	Short: "Kling AI commands",
	Long:  "Access Kling AI capabilities (Video generation, Voice cloning).",
}

func init() {
	Cmd.AddCommand(video.Cmd)
	Cmd.AddCommand(voice.Cmd)
}
