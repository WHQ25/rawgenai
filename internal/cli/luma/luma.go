package luma

import (
	"github.com/WHQ25/rawgenai/internal/cli/luma/image"
	"github.com/WHQ25/rawgenai/internal/cli/luma/video"
	"github.com/spf13/cobra"
)

// Cmd is the luma command
var Cmd = &cobra.Command{
	Use:   "luma",
	Short: "Luma AI commands",
	Long:  "Access Luma AI Dream Machine capabilities (Video, Image generation).",
}

func init() {
	Cmd.AddCommand(video.Cmd)
	Cmd.AddCommand(image.Cmd)
}
