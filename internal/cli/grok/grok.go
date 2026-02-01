package grok

import (
	"github.com/WHQ25/rawgenai/internal/cli/grok/video"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "grok",
	Short: "xAI Grok provider commands",
	Long:  "Commands for xAI Grok services including Image generation/editing and Video generation/editing.",
}

func init() {
	Cmd.AddCommand(imageCmd)
	Cmd.AddCommand(video.Cmd)
}
