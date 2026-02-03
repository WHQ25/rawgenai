package image

import "github.com/spf13/cobra"

// Cmd is the image subcommand
var Cmd = NewCmd()

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "image",
		Short: "Kling image generation commands",
		Long:  "Generate images using Kling image generation API.",
	}

	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newStatusCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newDownloadCmd())

	return cmd
}
