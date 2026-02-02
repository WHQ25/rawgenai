package tts

import (
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/cli/minimax/shared"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

const wsEndpoint = "wss://api.minimax.io/ws/v1/t2a_v2"

type wsMessage struct {
	Event string `json:"event,omitempty"`
	Data  struct {
		Audio string `json:"audio,omitempty"`
	} `json:"data,omitempty"`
	IsFinal bool `json:"is_final,omitempty"`
}

func runWebsocket(cmd *cobra.Command, text string, flags *ttsFlags) error {
	if flags.format != "mp3" && flags.speak {
		return common.WriteError(cmd, "invalid_format", "--speak only supports mp3 format")
	}

	apiKey := shared.GetMinimaxAPIKey()
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("MINIMAX_API_KEY"))
	}

	header := http.Header{}
	header.Set("Authorization", "Bearer "+apiKey)

	conn, _, err := websocket.DefaultDialer.Dial(wsEndpoint, header)
	if err != nil {
		return common.WriteError(cmd, "connection_error", fmt.Sprintf("cannot connect websocket: %s", err.Error()))
	}
	defer conn.Close()

	// Wait for connection ack if present
	var initMsg wsMessage
	_ = conn.ReadJSON(&initMsg)

	startMsg := map[string]any{
		"event": "task_start",
		"model": flags.model,
		"voice_setting": map[string]any{
			"voice_id": flags.voice,
			"speed":    flags.speed,
			"vol":      flags.vol,
			"pitch":    flags.pitch,
		},
		"audio_setting": map[string]any{
			"format": flags.format,
		},
		"continuous_sound": false,
	}
	if flags.sampleRate != 0 {
		startMsg["audio_setting"].(map[string]any)["sample_rate"] = flags.sampleRate
	}
	if flags.bitrate != 0 {
		startMsg["audio_setting"].(map[string]any)["bitrate"] = flags.bitrate
	}
	if flags.channel != 0 {
		startMsg["audio_setting"].(map[string]any)["channel"] = flags.channel
	}

	if err := conn.WriteJSON(startMsg); err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot send task_start: %s", err.Error()))
	}

	continueMsg := map[string]any{
		"event": "task_continue",
		"text":  text,
	}
	if err := conn.WriteJSON(continueMsg); err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot send task_continue: %s", err.Error()))
	}

	var outputPath string
	if flags.output != "" {
		absPath, err := filepath.Abs(flags.output)
		if err != nil {
			absPath = flags.output
		}
		outputPath = absPath
	}

	var outFile *os.File
	var errOpen error
	if outputPath != "" {
		outFile, errOpen = os.Create(outputPath)
		if errOpen != nil {
			return common.WriteError(cmd, "output_write_error", fmt.Sprintf("cannot create output file: %s", errOpen.Error()))
		}
		defer outFile.Close()
	}

	var writer io.Writer
	if outFile != nil {
		writer = outFile
	}

	var pipeWriter *io.PipeWriter
	if flags.speak {
		pr, pw := io.Pipe()
		pipeWriter = pw
		if writer != nil {
			writer = io.MultiWriter(writer, pw)
		} else {
			writer = pw
		}
		playErr := make(chan error, 1)
		go func() {
			playErr <- common.PlayMP3(pr)
		}()
		defer func() {
			pw.Close()
			_ = <-playErr
		}()
	}

	if writer == nil {
		return common.WriteError(cmd, "missing_output", "output file is required, use -o flag or --speak")
	}

	for {
		var msg wsMessage
		if err := conn.ReadJSON(&msg); err != nil {
			if pipeWriter != nil {
				pipeWriter.CloseWithError(err)
			}
			return common.WriteError(cmd, "stream_error", fmt.Sprintf("websocket read error: %s", err.Error()))
		}

		if msg.Data.Audio != "" {
			chunk, err := hex.DecodeString(msg.Data.Audio)
			if err != nil {
				return common.WriteError(cmd, "decode_error", fmt.Sprintf("cannot decode audio: %s", err.Error()))
			}
			if _, err := writer.Write(chunk); err != nil {
				return common.WriteError(cmd, "output_write_error", fmt.Sprintf("cannot write audio: %s", err.Error()))
			}
		}

		if msg.IsFinal {
			break
		}
	}

	_ = conn.WriteJSON(map[string]any{"event": "task_finish"})

	return common.WriteSuccess(cmd, ttsSyncResponse{
		Success: true,
		File:    outputPath,
		Model:   flags.model,
		Voice:   flags.voice,
	})
}
