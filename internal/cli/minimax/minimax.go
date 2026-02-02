package minimax

import (
	"github.com/WHQ25/rawgenai/internal/cli/minimax/image"
	"github.com/WHQ25/rawgenai/internal/cli/minimax/music"
	"github.com/WHQ25/rawgenai/internal/cli/minimax/tts"
	"github.com/WHQ25/rawgenai/internal/cli/minimax/video"
	"github.com/WHQ25/rawgenai/internal/cli/minimax/voice"
	"github.com/spf13/cobra"
)

// Cmd is the MiniMax command
var Cmd = &cobra.Command{
	Use:   "minimax",
	Short: "MiniMax AI commands",
	Long:  "Access MiniMax capabilities (Image, Video, TTS).",
}

func init() {
	Cmd.AddCommand(image.Cmd)
	Cmd.AddCommand(video.Cmd)
	Cmd.AddCommand(tts.Cmd)
	Cmd.AddCommand(voice.Cmd)
	Cmd.AddCommand(music.Cmd)
}
