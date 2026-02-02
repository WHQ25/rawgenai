package seed

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:   "seed",
	Short: "ByteDance Seed AI commands",
	Long:  "Access ByteDance Seed AI capabilities (TTS, Image, Video).",
}

func init() {
	Cmd.AddCommand(ttsCmd)
	Cmd.AddCommand(imageCmd)
	Cmd.AddCommand(videoCmd)
}
