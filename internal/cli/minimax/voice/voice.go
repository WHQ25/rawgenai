package voice

import "github.com/spf13/cobra"

// Cmd is the voice subcommand
var Cmd = NewCmd()

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "voice",
		Short: "Manage MiniMax voices",
		Long:  "List, clone, design, and delete MiniMax voices.",
	}

	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newUploadCmd())
	cmd.AddCommand(newCloneCmd())
	cmd.AddCommand(newDesignCmd())
	cmd.AddCommand(newDeleteCmd())

	return cmd
}
