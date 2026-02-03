package elevenlabs

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "elevenlabs",
	Short: "ElevenLabs provider commands",
	Long:  "Commands for ElevenLabs services including TTS, STT, Sound Effects, Music, Dialogue, and Voice Design.",
}

func init() {
	Cmd.AddCommand(ttsCmd)
	Cmd.AddCommand(sttCmd)
	Cmd.AddCommand(sfxCmd)
	Cmd.AddCommand(musicCmd)
	Cmd.AddCommand(dialogueCmd)
	Cmd.AddCommand(voiceCmd)
}
