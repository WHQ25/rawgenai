package image

import (
	"github.com/spf13/cobra"
)

// Valid image models
var validImageModels = map[string]bool{
	"photon-1":       true,
	"photon-flash-1": true,
}

// Valid aspect ratios
var validAspectRatios = map[string]bool{
	"1:1":  true,
	"16:9": true,
	"9:16": true,
	"4:3":  true,
	"3:4":  true,
	"21:9": true,
	"9:21": true,
}

// Valid image formats
var validImageFormats = map[string]bool{
	"jpg": true,
	"png": true,
}

// Cmd is the image subcommand
var Cmd = &cobra.Command{
	Use:   "image",
	Short: "Image generation commands",
	Long:  "Generate and manipulate images using Luma AI Photon models.",
}

func init() {
	Cmd.AddCommand(newCreateCmd())
	Cmd.AddCommand(newReframeCmd())
	Cmd.AddCommand(newStatusCmd())
	Cmd.AddCommand(newDownloadCmd())
	Cmd.AddCommand(newDeleteCmd())
}
