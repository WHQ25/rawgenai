package audio

import (
	"github.com/spf13/cobra"
)

// Cmd is the audio subcommand
var Cmd = &cobra.Command{
	Use:   "audio",
	Short: "Runway audio generation commands",
	Long:  "Generate audio using Runway AI (sound effects, TTS, speech-to-speech, dubbing, isolation).",
}

func init() {
	Cmd.AddCommand(newSfxCmd())
	Cmd.AddCommand(newTTSCmd())
	Cmd.AddCommand(newSTSCmd())
	Cmd.AddCommand(newDubbingCmd())
	Cmd.AddCommand(newIsolationCmd())
	Cmd.AddCommand(newStatusCmd())
	Cmd.AddCommand(newDownloadCmd())
	Cmd.AddCommand(newDeleteCmd())
}
