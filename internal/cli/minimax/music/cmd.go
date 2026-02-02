package music

import "github.com/spf13/cobra"

// Cmd is the music subcommand
var Cmd = &cobra.Command{
	Use:   "music",
	Short: "MiniMax music generation commands",
	Long:  "Generate music using MiniMax music generation API.",
}

func init() {
	Cmd.AddCommand(newCreateCmd())
}
