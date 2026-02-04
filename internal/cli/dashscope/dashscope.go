package dashscope

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:   "dashscope",
	Short: "DashScope (Tongyi Wanxiang) commands",
	Long:  "Access Alibaba Tongyi capabilities via DashScope API (Video, Image, TTS, STT).",
}

func init() {
	Cmd.AddCommand(videoCmd)
	Cmd.AddCommand(imageCmd)
	Cmd.AddCommand(ttsCmd)
	Cmd.AddCommand(sttCmd)
}
