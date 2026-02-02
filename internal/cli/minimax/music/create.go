package music

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/cli/minimax/shared"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

type createFlags struct {
	output     string
	lyricsFile string
	prompt     string
	stream     bool
	play       bool
	format     string
	sampleRate int
	bitrate    int
}

var validFormats = map[string]bool{
	"mp3": true,
	"wav": true,
	"pcm": true,
}

var validSampleRates = map[int]bool{
	16000: true,
	24000: true,
	32000: true,
	44100: true,
}

var validBitrates = map[int]bool{
	32000:  true,
	64000:  true,
	128000: true,
	256000: true,
}

func newCreateCmd() *cobra.Command {
	flags := &createFlags{}

	cmd := &cobra.Command{
		Use:           "create [lyrics]",
		Short:         "Generate music from lyrics",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.output, "output", "o", "", "Output file path")
	cmd.Flags().StringVar(&flags.lyricsFile, "lyrics-file", "", "Read lyrics from file")
	cmd.Flags().StringVarP(&flags.prompt, "prompt", "p", "", "Music style description (e.g., 'Pop, melancholic, rainy night')")
	cmd.Flags().BoolVar(&flags.stream, "stream", false, "Stream output to stdout")
	cmd.Flags().BoolVar(&flags.play, "play", false, "Play the generated music after creation")
	cmd.Flags().StringVarP(&flags.format, "format", "f", "mp3", "Audio format: mp3, wav, pcm")
	cmd.Flags().IntVar(&flags.sampleRate, "sample-rate", 44100, "Sample rate: 16000, 24000, 32000, 44100")
	cmd.Flags().IntVar(&flags.bitrate, "bitrate", 256000, "Bitrate: 32000, 64000, 128000, 256000")

	return cmd
}

type createResponse struct {
	Success  bool   `json:"success"`
	File     string `json:"file,omitempty"`
	Duration int    `json:"duration_ms,omitempty"`
	Size     int    `json:"size_bytes,omitempty"`
}

func runCreate(cmd *cobra.Command, args []string, flags *createFlags) error {
	lyrics, err := getLyrics(args, flags.lyricsFile, cmd.InOrStdin())
	if err != nil {
		return common.WriteError(cmd, "missing_lyrics", err.Error())
	}

	if len(lyrics) > 3500 {
		return common.WriteError(cmd, "lyrics_too_long", "lyrics must be 3500 characters or less")
	}

	if !flags.stream && flags.output == "" && !flags.play {
		return common.WriteError(cmd, "missing_output", "output file is required, use -o flag (or --play to play directly)")
	}

	if !validFormats[flags.format] {
		return common.WriteError(cmd, "invalid_format", "format must be mp3, wav, or pcm")
	}

	if !validSampleRates[flags.sampleRate] {
		return common.WriteError(cmd, "invalid_sample_rate", "sample-rate must be 16000, 24000, 32000, or 44100")
	}

	if !validBitrates[flags.bitrate] {
		return common.WriteError(cmd, "invalid_bitrate", "bitrate must be 32000, 64000, 128000, or 256000")
	}

	if flags.output != "" {
		ext := strings.ToLower(filepath.Ext(flags.output))
		expectedExt := "." + flags.format
		if ext != expectedExt {
			return common.WriteError(cmd, "format_mismatch", fmt.Sprintf("output file extension should be %s for format %s", expectedExt, flags.format))
		}
	}

	apiKey := shared.GetMinimaxAPIKey()
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("MINIMAX_API_KEY"))
	}

	body := map[string]any{
		"model":  "music-2.5",
		"lyrics": lyrics,
		"stream": flags.stream,
		"audio_setting": map[string]any{
			"format":      flags.format,
			"sample_rate": flags.sampleRate,
			"bitrate":     flags.bitrate,
		},
	}

	if flags.prompt != "" {
		body["prompt"] = flags.prompt
	}

	// Use hex for stream mode, url for non-stream (easier to download)
	if flags.stream {
		body["output_format"] = "hex"
	} else {
		body["output_format"] = "url"
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot serialize request: %s", err.Error()))
	}

	req, err := shared.CreateRequest("POST", "/v1/music_generation", bytes.NewReader(jsonBody))
	if err != nil {
		return common.WriteError(cmd, "request_error", err.Error())
	}

	if flags.stream {
		return handleStreamResponse(cmd, req, flags)
	}
	return handleSyncResponse(cmd, req, flags)
}

func handleSyncResponse(cmd *cobra.Command, req *http.Request, flags *createFlags) error {
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return common.WriteError(cmd, "request_error", err.Error())
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot read response: %s", err.Error()))
	}

	if resp.StatusCode != http.StatusOK {
		return common.WriteError(cmd, "api_error", fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(respBody)))
	}

	var apiResp struct {
		Data struct {
			Audio  string `json:"audio"`
			Status int    `json:"status"`
		} `json:"data"`
		ExtraInfo struct {
			MusicDuration int `json:"music_duration"`
			MusicSize     int `json:"music_size"`
		} `json:"extra_info"`
		BaseResp struct {
			StatusCode int    `json:"status_code"`
			StatusMsg  string `json:"status_msg"`
		} `json:"base_resp"`
	}
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot parse response: %s", err.Error()))
	}

	if apiResp.BaseResp.StatusCode != 0 {
		return common.WriteError(cmd, "api_error", fmt.Sprintf("api error %d: %s", apiResp.BaseResp.StatusCode, apiResp.BaseResp.StatusMsg))
	}

	// Download from URL
	audioURL := apiResp.Data.Audio
	if audioURL == "" {
		return common.WriteError(cmd, "no_audio", "no audio URL in response")
	}

	audioResp, err := http.Get(audioURL)
	if err != nil {
		return common.WriteError(cmd, "download_error", fmt.Sprintf("cannot download audio: %s", err.Error()))
	}
	defer audioResp.Body.Close()

	if audioResp.StatusCode != http.StatusOK {
		return common.WriteError(cmd, "download_error", fmt.Sprintf("download failed with status %d", audioResp.StatusCode))
	}

	audioData, err := io.ReadAll(audioResp.Body)
	if err != nil {
		return common.WriteError(cmd, "download_error", fmt.Sprintf("cannot read audio: %s", err.Error()))
	}

	// Determine output path
	var absPath string
	var useTempFile bool

	if flags.output != "" {
		absPath, err = filepath.Abs(flags.output)
		if err != nil {
			absPath = flags.output
		}
	} else {
		// Create temp file for --play only mode
		tmpFile, err := os.CreateTemp("", "minimax-music-*."+flags.format)
		if err != nil {
			return common.WriteError(cmd, "temp_file_error", fmt.Sprintf("cannot create temp file: %s", err.Error()))
		}
		absPath = tmpFile.Name()
		tmpFile.Close()
		useTempFile = true
	}

	if err := os.WriteFile(absPath, audioData, 0644); err != nil {
		return common.WriteError(cmd, "output_write_error", fmt.Sprintf("cannot write output file: %s", err.Error()))
	}

	result := createResponse{
		Success:  true,
		Duration: apiResp.ExtraInfo.MusicDuration,
		Size:     len(audioData),
	}

	if !useTempFile {
		result.File = absPath
	}

	if flags.play {
		if err := common.PlayFile(absPath); err != nil {
			return common.WriteError(cmd, "playback_error", fmt.Sprintf("cannot play audio: %s", err.Error()))
		}
	}

	if useTempFile {
		os.Remove(absPath)
	}

	return common.WriteSuccess(cmd, result)
}

func handleStreamResponse(cmd *cobra.Command, req *http.Request, flags *createFlags) error {
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return common.WriteError(cmd, "request_error", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return common.WriteError(cmd, "api_error", fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)))
	}

	reader := bufio.NewReader(resp.Body)
	var audioBuffer bytes.Buffer
	var totalBytes int

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return common.WriteError(cmd, "stream_error", fmt.Sprintf("error reading stream: %s", err.Error()))
		}

		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		// Parse SSE data
		if bytes.HasPrefix(line, []byte("data:")) {
			data := bytes.TrimPrefix(line, []byte("data:"))
			data = bytes.TrimSpace(data)

			var chunk struct {
				Data struct {
					Audio  string `json:"audio"`
					Status int    `json:"status"`
				} `json:"data"`
				BaseResp struct {
					StatusCode int    `json:"status_code"`
					StatusMsg  string `json:"status_msg"`
				} `json:"base_resp"`
			}

			if err := json.Unmarshal(data, &chunk); err != nil {
				continue
			}

			if chunk.BaseResp.StatusCode != 0 {
				return common.WriteError(cmd, "api_error", fmt.Sprintf("api error %d: %s", chunk.BaseResp.StatusCode, chunk.BaseResp.StatusMsg))
			}

			if chunk.Data.Audio != "" {
				audioBytes, err := hex.DecodeString(chunk.Data.Audio)
				if err != nil {
					continue
				}
				totalBytes += len(audioBytes)
				if flags.play {
					// Collect for playback
					audioBuffer.Write(audioBytes)
				} else {
					// Stream to stdout
					os.Stdout.Write(audioBytes)
				}
			}

			if chunk.Data.Status == 2 {
				break
			}
		}
	}

	if flags.play {
		// Save to temp file and play
		tmpFile, err := os.CreateTemp("", "minimax-music-*."+flags.format)
		if err != nil {
			return common.WriteError(cmd, "temp_file_error", fmt.Sprintf("cannot create temp file: %s", err.Error()))
		}
		tmpPath := tmpFile.Name()
		tmpFile.Write(audioBuffer.Bytes())
		tmpFile.Close()

		if err := common.PlayFile(tmpPath); err != nil {
			os.Remove(tmpPath)
			return common.WriteError(cmd, "playback_error", fmt.Sprintf("cannot play audio: %s", err.Error()))
		}
		os.Remove(tmpPath)

		return common.WriteSuccess(cmd, createResponse{
			Success: true,
			Size:    totalBytes,
		})
	}

	// Non-play stream mode: completion message to stderr (audio already streamed to stdout)
	fmt.Fprintf(os.Stderr, `{"success":true,"size_bytes":%d}`+"\n", totalBytes)
	return nil
}

func getLyrics(args []string, filePath string, stdin io.Reader) (string, error) {
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
				return "", errors.New("no lyrics provided")
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

	return "", errors.New("no lyrics provided")
}
