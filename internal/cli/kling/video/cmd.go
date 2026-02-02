package video

import "github.com/spf13/cobra"

var Cmd = NewCmd()

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "video",
		Short: "Generate videos using Kling Omni-Video O1",
		Long:  "Generate videos using Kling Omni-Video O1 model.",
	}

	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newText2VideoCmd())
	cmd.AddCommand(newImage2VideoCmd())
	cmd.AddCommand(newMotionControlCmd())
	cmd.AddCommand(newAvatarCmd())
	cmd.AddCommand(newStatusCmd())
	cmd.AddCommand(newDownloadCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newExtendCmd())
	cmd.AddCommand(newAddSoundCmd())
	cmd.AddCommand(newElementCmd())

	return cmd
}
