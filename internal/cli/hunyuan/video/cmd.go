package video

import "github.com/spf13/cobra"

// Cmd is the video subcommand
var Cmd = NewCmd()

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "video",
		Short: "Hunyuan video generation commands",
		Long:  "Generate videos using Hunyuan video generation API.",
	}

	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newStatusCmd())
	cmd.AddCommand(newDownloadCmd())

	return cmd
}
