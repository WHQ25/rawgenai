package image

import (
	"github.com/spf13/cobra"
)

// Cmd is the image subcommand
var Cmd = &cobra.Command{
	Use:   "image",
	Short: "Runway image generation commands",
	Long:  "Generate images using Runway AI models.",
}

func init() {
	Cmd.AddCommand(newCreateCmd())
	Cmd.AddCommand(newStatusCmd())
	Cmd.AddCommand(newDownloadCmd())
	Cmd.AddCommand(newDeleteCmd())
}
