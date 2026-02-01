package google

import (
	"github.com/WHQ25/rawgenai/internal/cli/google/video"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "google",
	Short: "Google Gemini provider commands",
	Long:  "Commands for Google Gemini services including TTS, STT, Image and Video generation.",
}

func init() {
	Cmd.AddCommand(imageCmd)
	Cmd.AddCommand(video.Cmd)
	Cmd.AddCommand(ttsCmd)
	Cmd.AddCommand(sttCmd)
}
