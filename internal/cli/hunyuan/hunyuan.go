package hunyuan

import (
	"github.com/WHQ25/rawgenai/internal/cli/hunyuan/image"
	"github.com/WHQ25/rawgenai/internal/cli/hunyuan/video"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "hunyuan",
	Short: "Hunyuan commands",
	Long:  "Access Tencent Hunyuan capabilities (Image, Video generation).",
}

func init() {
	Cmd.AddCommand(image.Cmd)
	Cmd.AddCommand(video.Cmd)
}
