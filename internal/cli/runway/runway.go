package runway

import (
	"github.com/WHQ25/rawgenai/internal/cli/runway/audio"
	"github.com/WHQ25/rawgenai/internal/cli/runway/image"
	"github.com/WHQ25/rawgenai/internal/cli/runway/video"
	"github.com/spf13/cobra"
)

// Cmd is the runway command
var Cmd = &cobra.Command{
	Use:   "runway",
	Short: "Runway AI commands",
	Long:  "Access Runway AI capabilities (Video, Image, Audio generation).",
}

func init() {
	Cmd.AddCommand(video.Cmd)
	Cmd.AddCommand(image.Cmd)
	Cmd.AddCommand(audio.Cmd)
}
