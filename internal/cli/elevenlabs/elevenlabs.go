package elevenlabs

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "elevenlabs",
	Short: "ElevenLabs provider commands",
	Long:  "Commands for ElevenLabs services including TTS, STT, and Sound Effects generation.",
}

func init() {
	Cmd.AddCommand(ttsCmd)
	Cmd.AddCommand(sttCmd)
	Cmd.AddCommand(sfxCmd)
}
