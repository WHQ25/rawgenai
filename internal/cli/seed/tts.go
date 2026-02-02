package seed

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

// V3 bidirectional stream endpoint
const ttsEndpoint = "wss://openspeech.bytedance.com/api/v3/tts/bidirection"

// Event types
const (
	EventStartConnection  int32 = 1
	EventFinishConnection int32 = 2
	EventConnectionStarted int32 = 50
	EventConnectionFailed  int32 = 51
	EventStartSession      int32 = 100
	EventFinishSession     int32 = 102
	EventSessionStarted    int32 = 150
	EventSessionFinished   int32 = 152
	EventSessionFailed     int32 = 153
	EventTaskRequest       int32 = 200
	EventTTSSentenceStart  int32 = 350
	EventTTSSentenceEnd    int32 = 351
	EventTTSResponse       int32 = 352
)

// Message types
const (
	MsgTypeFullClientRequest  uint8 = 0b0001
	MsgTypeFullServerResponse uint8 = 0b1001
	MsgTypeAudioOnlyResponse  uint8 = 0b1011
	MsgTypeError              uint8 = 0b1111
)

// Flags
const (
	FlagWithEvent uint8 = 0b0100
)

type ttsFlags struct {
	output     string
	promptFile string
	voice      string
	format     string
	sampleRate int
	speed      int
	volume     int
	speak      bool
}

type ttsResponse struct {
	Success bool   `json:"success"`
	File    string `json:"file,omitempty"`
	Voice   string `json:"voice,omitempty"`
}

var ttsCmd = newTTSCmd()

func newTTSCmd() *cobra.Command {
	flags := &ttsFlags{}

	cmd := &cobra.Command{
		Use:           "tts [text]",
		Short:         "Text to Speech using Seed TTS models",
		Long:          "Convert text to speech using ByteDance Seed TTS models (V3 bidirectional streaming).",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTTS(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.output, "output", "o", "", "Output file path")
	cmd.Flags().StringVar(&flags.promptFile, "prompt-file", "", "Read text from file")
	cmd.Flags().StringVarP(&flags.voice, "voice", "V", "zh_female_vv_uranus_bigtts", "Voice name")
	cmd.Flags().StringVar(&flags.format, "format", "mp3", "Audio format: mp3, pcm, ogg_opus")
	cmd.Flags().IntVar(&flags.sampleRate, "sample-rate", 24000, "Sample rate: 8000, 16000, 24000, etc.")
	cmd.Flags().IntVar(&flags.speed, "speed", 0, "Speech rate: -50 to 100 (0 = normal)")
	cmd.Flags().IntVar(&flags.volume, "volume", 0, "Volume: -50 to 100 (0 = normal)")
	cmd.Flags().BoolVar(&flags.speak, "speak", false, "Play audio after generation")

	return cmd
}

func runTTS(cmd *cobra.Command, args []string, flags *ttsFlags) error {
	// Get text from args, file, or stdin
	text, err := getText(args, flags.promptFile, cmd.InOrStdin())
	if err != nil {
		return common.WriteError(cmd, "missing_text", err.Error())
	}

	// Validate output or speak
	if flags.output == "" && !flags.speak {
		return common.WriteError(cmd, "missing_output", "output file is required, use -o flag or --speak")
	}

	// Validate format
	validFormats := map[string]bool{"mp3": true, "pcm": true, "ogg_opus": true}
	if !validFormats[flags.format] {
		return common.WriteError(cmd, "invalid_format", "format must be mp3, pcm, or ogg_opus")
	}

	// Streaming playback only supports mp3
	if flags.speak && flags.format != "mp3" {
		return common.WriteError(cmd, "invalid_format", "--speak only supports mp3 format")
	}

	// Validate speed
	if flags.speed < -50 || flags.speed > 100 {
		return common.WriteError(cmd, "invalid_speed", "speed must be between -50 and 100")
	}

	// Validate volume
	if flags.volume < -50 || flags.volume > 100 {
		return common.WriteError(cmd, "invalid_volume", "volume must be between -50 and 100")
	}

	// Get credentials
	appID := config.GetAPIKey("SEED_APP_ID")
	if appID == "" {
		return common.WriteError(cmd, "missing_credentials", config.GetMissingKeyMessage("SEED_APP_ID"))
	}
	accessToken := config.GetAPIKey("SEED_ACCESS_TOKEN")
	if accessToken == "" {
		return common.WriteError(cmd, "missing_credentials", config.GetMissingKeyMessage("SEED_ACCESS_TOKEN"))
	}

	// Determine output path
	var absPath string
	if flags.output != "" {
		absPath, err = filepath.Abs(flags.output)
		if err != nil {
			absPath = flags.output
		}
	}

	// Stream mode: --speak (with optional --output)
	if flags.speak {
		return runStreamTTS(cmd, appID, accessToken, text, absPath, flags)
	}

	// File mode: --output only
	return runFileTTS(cmd, appID, accessToken, text, absPath, flags)
}

func runStreamTTS(cmd *cobra.Command, appID, accessToken, text, outputPath string, flags *ttsFlags) error {
	// Create pipe for streaming playback
	pr, pw := io.Pipe()

	// Create writers: always include pipe for playback
	var writers []io.Writer
	writers = append(writers, pw)

	// Optionally write to file
	var outFile *os.File
	if outputPath != "" {
		var err error
		outFile, err = os.Create(outputPath)
		if err != nil {
			pw.Close()
			return common.WriteError(cmd, "output_write_error", fmt.Sprintf("cannot create output file: %s", err.Error()))
		}
		defer outFile.Close()
		writers = append(writers, outFile)
	}

	// Multi-writer for all destinations
	mw := io.MultiWriter(writers...)

	// Start playback in goroutine
	playErr := make(chan error, 1)
	go func() {
		playErr <- common.PlayMP3(pr)
	}()

	// Stream audio to writers
	if err := streamAudio(context.Background(), appID, accessToken, text, flags, mw); err != nil {
		pw.CloseWithError(err)
		return common.WriteError(cmd, "api_error", err.Error())
	}
	pw.Close()

	// Wait for playback to finish
	if err := <-playErr; err != nil {
		return common.WriteError(cmd, "playback_error", fmt.Sprintf("cannot play audio: %s", err.Error()))
	}

	// Return success
	result := ttsResponse{
		Success: true,
		File:    outputPath,
		Voice:   flags.voice,
	}
	return common.WriteSuccess(cmd, result)
}

func runFileTTS(cmd *cobra.Command, appID, accessToken, text, outputPath string, flags *ttsFlags) error {
	// Create output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return common.WriteError(cmd, "output_write_error", fmt.Sprintf("cannot create output file: %s", err.Error()))
	}
	defer outFile.Close()

	// Stream audio to file
	if err := streamAudio(context.Background(), appID, accessToken, text, flags, outFile); err != nil {
		os.Remove(outputPath)
		return common.WriteError(cmd, "api_error", err.Error())
	}

	// Return success
	result := ttsResponse{
		Success: true,
		File:    outputPath,
		Voice:   flags.voice,
	}
	return common.WriteSuccess(cmd, result)
}

func streamAudio(ctx context.Context, appID, accessToken, text string, flags *ttsFlags, w io.Writer) error {
	// Setup WebSocket connection headers
	header := http.Header{}
	header.Set("X-Api-App-Key", appID)
	header.Set("X-Api-Access-Key", accessToken)
	header.Set("X-Api-Resource-Id", "seed-tts-2.0")
	header.Set("X-Api-Connect-Id", uuid.New().String())

	// Connect
	conn, resp, err := websocket.DefaultDialer.DialContext(ctx, ttsEndpoint, header)
	if err != nil {
		if resp != nil {
			// Read response body for error details
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("connection failed (status %d): %s", resp.StatusCode, string(body))
		}
		return fmt.Errorf("connection failed: %w", err)
	}
	defer conn.Close()

	sessionID := uuid.New().String()

	// 1. Send StartConnection
	if err := sendEvent(conn, EventStartConnection, "", nil); err != nil {
		return fmt.Errorf("StartConnection failed: %w", err)
	}

	// 2. Wait for ConnectionStarted
	if err := waitForEvent(conn, EventConnectionStarted); err != nil {
		return fmt.Errorf("ConnectionStarted failed: %w", err)
	}

	// 3. Send StartSession with config
	sessionPayload := buildSessionPayload(text, flags)
	if err := sendEvent(conn, EventStartSession, sessionID, sessionPayload); err != nil {
		return fmt.Errorf("StartSession failed: %w", err)
	}

	// 4. Wait for SessionStarted
	if err := waitForEvent(conn, EventSessionStarted); err != nil {
		return fmt.Errorf("SessionStarted failed: %w", err)
	}

	// 5. Send TaskRequest with text
	taskPayload := map[string]any{
		"req_params": map[string]any{
			"text": text,
		},
	}
	taskBytes, _ := json.Marshal(taskPayload)
	if err := sendEvent(conn, EventTaskRequest, sessionID, taskBytes); err != nil {
		return fmt.Errorf("TaskRequest failed: %w", err)
	}

	// 6. Send FinishSession
	if err := sendEvent(conn, EventFinishSession, sessionID, nil); err != nil {
		return fmt.Errorf("FinishSession failed: %w", err)
	}

	// 7. Receive audio chunks until SessionFinished, write to output
	for {
		msgType, eventType, payload, err := receiveMessage(conn)
		if err != nil {
			return fmt.Errorf("receive failed: %w", err)
		}

		switch {
		case msgType == MsgTypeAudioOnlyResponse && eventType == EventTTSResponse:
			if _, err := w.Write(payload); err != nil {
				return fmt.Errorf("write failed: %w", err)
			}
		case msgType == MsgTypeFullServerResponse && eventType == EventSessionFinished:
			// 8. Send FinishConnection
			sendEvent(conn, EventFinishConnection, "", nil)
			return nil
		case msgType == MsgTypeFullServerResponse && eventType == EventSessionFailed:
			return fmt.Errorf("session failed: %s", string(payload))
		case msgType == MsgTypeError:
			return fmt.Errorf("server error: %s", string(payload))
		}
	}
}

func buildSessionPayload(text string, flags *ttsFlags) []byte {
	payload := map[string]any{
		"user": map[string]any{
			"uid": uuid.New().String(),
		},
		"req_params": map[string]any{
			"text":    text,
			"speaker": flags.voice,
			"audio_params": map[string]any{
				"format":       flags.format,
				"sample_rate":  flags.sampleRate,
				"speech_rate":  flags.speed,
				"loudness_rate": flags.volume,
			},
		},
	}
	data, _ := json.Marshal(payload)
	return data
}

// Binary protocol helpers

func sendEvent(conn *websocket.Conn, eventType int32, sessionID string, payload []byte) error {
	if payload == nil {
		payload = []byte("{}")
	}

	buf := new(bytes.Buffer)

	// Header (4 bytes)
	buf.WriteByte(0x11)                           // version=1, header_size=1
	buf.WriteByte(MsgTypeFullClientRequest<<4 | FlagWithEvent) // msg_type + flags
	buf.WriteByte(0x10)                           // serialization=JSON, compression=none
	buf.WriteByte(0x00)                           // reserved

	// Event type (4 bytes, big endian)
	binary.Write(buf, binary.BigEndian, eventType)

	// Session ID (if not connection-level event)
	if eventType != EventStartConnection && eventType != EventFinishConnection {
		binary.Write(buf, binary.BigEndian, uint32(len(sessionID)))
		buf.WriteString(sessionID)
	}

	// Payload
	binary.Write(buf, binary.BigEndian, uint32(len(payload)))
	buf.Write(payload)

	return conn.WriteMessage(websocket.BinaryMessage, buf.Bytes())
}

func waitForEvent(conn *websocket.Conn, expectedEvent int32) error {
	msgType, eventType, payload, err := receiveMessage(conn)
	if err != nil {
		return err
	}

	if msgType == MsgTypeError {
		return fmt.Errorf("server error: %s", string(payload))
	}

	if eventType != expectedEvent {
		return fmt.Errorf("unexpected event: got %d, expected %d", eventType, expectedEvent)
	}

	// Check for failure events
	if eventType == EventConnectionFailed || eventType == EventSessionFailed {
		return fmt.Errorf("operation failed: %s", string(payload))
	}

	return nil
}

func receiveMessage(conn *websocket.Conn) (msgType uint8, eventType int32, payload []byte, err error) {
	_, frame, err := conn.ReadMessage()
	if err != nil {
		return 0, 0, nil, err
	}

	if len(frame) < 4 {
		return 0, 0, nil, errors.New("frame too short")
	}

	msgType = (frame[1] >> 4) & 0x0F
	flags := frame[1] & 0x0F
	hasEvent := (flags & FlagWithEvent) != 0

	pos := 4 // after header

	// For error messages, read error code first
	if msgType == MsgTypeError {
		if len(frame) >= pos+4 {
			// Error code is 4 bytes
			pos += 4
		}
		// Rest is payload
		if len(frame) > pos {
			payload = frame[pos:]
		}
		return msgType, 0, payload, nil
	}

	// Read event if present
	if hasEvent && len(frame) >= pos+4 {
		eventType = int32(binary.BigEndian.Uint32(frame[pos : pos+4]))
		pos += 4

		// Skip session ID for session-level events
		if eventType != EventConnectionStarted && eventType != EventConnectionFailed &&
		   eventType >= 100 && len(frame) >= pos+4 {
			sidLen := binary.BigEndian.Uint32(frame[pos : pos+4])
			pos += 4 + int(sidLen)
		}

		// Skip connection ID for connection events
		if (eventType == EventConnectionStarted || eventType == EventConnectionFailed) && len(frame) >= pos+4 {
			cidLen := binary.BigEndian.Uint32(frame[pos : pos+4])
			pos += 4 + int(cidLen)
		}
	}

	// Read payload
	if len(frame) >= pos+4 {
		payloadLen := binary.BigEndian.Uint32(frame[pos : pos+4])
		pos += 4
		if len(frame) >= pos+int(payloadLen) {
			payload = frame[pos : pos+int(payloadLen)]
		}
	}

	return msgType, eventType, payload, nil
}

func getText(args []string, filePath string, stdin io.Reader) (string, error) {
	// From positional argument
	if len(args) > 0 {
		text := strings.TrimSpace(strings.Join(args, " "))
		if text != "" {
			return text, nil
		}
	}

	// From file
	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("cannot read file: %w", err)
		}
		text := strings.TrimSpace(string(data))
		if text != "" {
			return text, nil
		}
	}

	// From stdin
	if stdin != nil {
		if f, ok := stdin.(*os.File); ok {
			stat, _ := f.Stat()
			if (stat.Mode() & os.ModeCharDevice) != 0 {
				return "", errors.New("no text provided, use positional argument, --prompt-file, or pipe from stdin")
			}
		}
		data, err := io.ReadAll(stdin)
		if err != nil {
			return "", fmt.Errorf("cannot read stdin: %w", err)
		}
		text := strings.TrimSpace(string(data))
		if text != "" {
			return text, nil
		}
	}

	return "", errors.New("no text provided, use positional argument, --prompt-file, or pipe from stdin")
}
