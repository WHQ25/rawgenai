package video

import (
	"github.com/spf13/cobra"
)

// Cmd is the video subcommand
var Cmd = &cobra.Command{
	Use:   "video",
	Short: "Runway video generation commands",
	Long:  "Generate videos using Runway AI models (image-to-video, text-to-video, video-to-video, upscale, character performance).",
}

func init() {
	Cmd.AddCommand(newImage2VideoCmd())
	Cmd.AddCommand(newText2VideoCmd())
	Cmd.AddCommand(newVideo2VideoCmd())
	Cmd.AddCommand(newUpscaleCmd())
	Cmd.AddCommand(newCharacterCmd())
	Cmd.AddCommand(newStatusCmd())
	Cmd.AddCommand(newDownloadCmd())
	Cmd.AddCommand(newDeleteCmd())
}
