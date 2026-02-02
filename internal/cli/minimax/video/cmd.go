package video

import "github.com/spf13/cobra"

// Cmd is the video subcommand
var Cmd = &cobra.Command{
	Use:   "video",
	Short: "MiniMax video generation commands",
	Long:  "Generate and manage videos using MiniMax video generation API.",
}

func init() {
	Cmd.AddCommand(newCreateCmd())
	Cmd.AddCommand(newStatusCmd())
	Cmd.AddCommand(newDownloadCmd())
}
