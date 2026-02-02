package tts

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/spf13/cobra"
)

// Cmd is the tts subcommand
var Cmd = newTTSCmd()

type ttsFlags struct {
	output     string
	promptFile string
	model      string
	voice      string
	speed      float64
	vol        float64
	pitch      int
	format     string
	sampleRate int
	bitrate    int
	channel    int
	stream     bool
	speak      bool
}

func newTTSCmd() *cobra.Command {
	flags := &ttsFlags{}

	cmd := &cobra.Command{
		Use:           "tts [text]",
		Short:         "MiniMax text-to-speech",
		Long:          "Synchronous TTS over HTTP by default. Use --stream for WebSocket streaming.",
		Args:          cobra.ArbitraryArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSync(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.output, "output", "o", "", "Output file path")
	cmd.Flags().StringVar(&flags.promptFile, "prompt-file", "", "Read text from file")
	cmd.Flags().StringVarP(&flags.model, "model", "m", "speech-2.8-hd", "Model name")
	cmd.Flags().StringVar(&flags.voice, "voice", "English_Graceful_Lady", "Voice ID")
	cmd.Flags().Float64Var(&flags.speed, "speed", 1, "Speech speed (0.5-2.0)")
	cmd.Flags().Float64Var(&flags.vol, "vol", 1, "Speech volume (0-10]")
	cmd.Flags().IntVar(&flags.pitch, "pitch", 0, "Speech pitch (-12 to 12)")
	cmd.Flags().StringVar(&flags.format, "format", "mp3", "Audio format: mp3, pcm, flac, wav")
	cmd.Flags().IntVar(&flags.sampleRate, "sample-rate", 0, "Sample rate (optional)")
	cmd.Flags().IntVar(&flags.bitrate, "bitrate", 0, "Bitrate (mp3 only)")
	cmd.Flags().IntVar(&flags.channel, "channel", 0, "Channel count (1 or 2)")
	cmd.Flags().BoolVar(&flags.stream, "stream", false, "Use WebSocket streaming")
	cmd.Flags().BoolVar(&flags.speak, "speak", false, "Play audio after generation")

	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newStatusCmd())
	cmd.AddCommand(newDownloadCmd())

	return cmd
}

func getText(args []string, filePath string, stdin io.Reader) (string, error) {
	if len(args) > 0 {
		text := strings.TrimSpace(strings.Join(args, " "))
		if text != "" {
			return text, nil
		}
	}

	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("cannot read file: %s", err.Error())
		}
		text := strings.TrimSpace(string(data))
		if text != "" {
			return text, nil
		}
	}

	if stdin != nil {
		if f, ok := stdin.(*os.File); ok {
			stat, _ := f.Stat()
			if (stat.Mode() & os.ModeCharDevice) != 0 {
				return "", errors.New("no text provided")
			}
		}
		data, err := io.ReadAll(stdin)
		if err != nil {
			return "", fmt.Errorf("cannot read stdin: %s", err.Error())
		}
		text := strings.TrimSpace(string(data))
		if text != "" {
			return text, nil
		}
	}

	return "", errors.New("no text provided")
}

func validateSyncFlags(cmd *cobra.Command, flags *ttsFlags) error {
	if flags.output == "" && !flags.speak {
		return common.WriteError(cmd, "missing_output", "output file is required, use -o flag or --speak")
	}

	if !validFormats[flags.format] {
		return common.WriteError(cmd, "invalid_format", "format must be mp3, pcm, flac, or wav")
	}

	if flags.speak && flags.format != "mp3" {
		return common.WriteError(cmd, "invalid_format", "--speak only supports mp3 format")
	}

	if flags.speed < 0.5 || flags.speed > 2 {
		return common.WriteError(cmd, "invalid_speed", "speed must be between 0.5 and 2.0")
	}
	if flags.vol <= 0 || flags.vol > 10 {
		return common.WriteError(cmd, "invalid_volume", "vol must be in (0, 10]")
	}
	if flags.pitch < -12 || flags.pitch > 12 {
		return common.WriteError(cmd, "invalid_pitch", "pitch must be between -12 and 12")
	}

	if flags.sampleRate != 0 && !validSampleRates[flags.sampleRate] {
		return common.WriteError(cmd, "invalid_sample_rate", "invalid sample rate")
	}
	if flags.bitrate != 0 && !validBitrates[flags.bitrate] {
		return common.WriteError(cmd, "invalid_bitrate", "invalid bitrate")
	}
	if flags.channel != 0 && flags.channel != 1 && flags.channel != 2 {
		return common.WriteError(cmd, "invalid_channel", "channel must be 1 or 2")
	}

	return nil
}
