package openai

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "openai",
	Short: "OpenAI provider commands",
	Long:  "Commands for OpenAI services including TTS, STT, and Image generation.",
}

func init() {
	Cmd.AddCommand(ttsCmd)
	Cmd.AddCommand(imageCmd)
}
