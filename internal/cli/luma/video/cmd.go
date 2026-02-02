package video

import "github.com/spf13/cobra"

// Cmd is the video subcommand
var Cmd = &cobra.Command{
	Use:   "video",
	Short: "Luma video generation commands",
	Long:  "Generate and manage videos using Luma Dream Machine.",
}

func init() {
	Cmd.AddCommand(newCreateCmd())
	Cmd.AddCommand(newExtendCmd())
	Cmd.AddCommand(newUpscaleCmd())
	Cmd.AddCommand(newAudioCmd())
	Cmd.AddCommand(newModifyCmd())
	Cmd.AddCommand(newStatusCmd())
	Cmd.AddCommand(newDownloadCmd())
	Cmd.AddCommand(newDeleteCmd())
	Cmd.AddCommand(newListCmd())
}
