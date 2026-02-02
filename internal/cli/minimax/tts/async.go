package tts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/cli/minimax/shared"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

type createFlags struct {
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
	fileID     int64
}

func newCreateCmd() *cobra.Command {
	flags := &createFlags{}

	cmd := &cobra.Command{
		Use:           "create [text]",
		Short:         "Create async TTS task (long text)",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(cmd, args, flags)
		},
	}

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
	cmd.Flags().Int64Var(&flags.fileID, "file-id", 0, "Text file ID for async synthesis")

	return cmd
}

type createResponse struct {
	Success         bool   `json:"success"`
	TaskID          int64  `json:"task_id,omitempty"`
	TaskToken       string `json:"task_token,omitempty"`
	FileID          int64  `json:"file_id,omitempty"`
	UsageCharacters int64  `json:"usage_characters,omitempty"`
}

func runCreate(cmd *cobra.Command, args []string, flags *createFlags) error {
	text, err := getTextOptional(args, flags.promptFile, cmd.InOrStdin())
	if err != nil {
		return common.WriteError(cmd, "missing_text", err.Error())
	}

	if flags.fileID == 0 && strings.TrimSpace(text) == "" {
		return common.WriteError(cmd, "missing_text", "text or --file-id is required")
	}
	if flags.fileID != 0 && strings.TrimSpace(text) != "" {
		return common.WriteError(cmd, "invalid_parameter", "text and --file-id are mutually exclusive")
	}

	if !validFormats[flags.format] {
		return common.WriteError(cmd, "invalid_format", "format must be mp3, pcm, flac, or wav")
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

	apiKey := shared.GetMinimaxAPIKey()
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("MINIMAX_API_KEY"))
	}

	body := map[string]any{
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
	}

	if flags.fileID != 0 {
		body["text_file_id"] = flags.fileID
	} else {
		body["text"] = text
	}

	if flags.sampleRate != 0 {
		body["audio_setting"].(map[string]any)["audio_sample_rate"] = flags.sampleRate
	}
	if flags.bitrate != 0 {
		body["audio_setting"].(map[string]any)["bitrate"] = flags.bitrate
	}
	if flags.channel != 0 {
		body["audio_setting"].(map[string]any)["channel"] = flags.channel
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot serialize request: %s", err.Error()))
	}

	req, err := shared.CreateRequest("POST", "/v1/t2a_async_v2", bytes.NewReader(jsonBody))
	if err != nil {
		return common.WriteError(cmd, "request_error", err.Error())
	}

	resp, err := shared.DoRequest(req)
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
		TaskID          int64  `json:"task_id"`
		TaskToken       string `json:"task_token"`
		FileID          int64  `json:"file_id"`
		UsageCharacters int64  `json:"usage_characters"`
		BaseResp        struct {
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

	return common.WriteSuccess(cmd, createResponse{
		Success:         true,
		TaskID:          apiResp.TaskID,
		TaskToken:       apiResp.TaskToken,
		FileID:          apiResp.FileID,
		UsageCharacters: apiResp.UsageCharacters,
	})
}

func newStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "status <task_id>",
		Short:         "Query async TTS task status",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd, args[0])
		},
	}
	return cmd
}

func runStatus(cmd *cobra.Command, taskID string) error {
	if strings.TrimSpace(taskID) == "" {
		return common.WriteError(cmd, "missing_task_id", "task_id is required")
	}

	apiKey := shared.GetMinimaxAPIKey()
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("MINIMAX_API_KEY"))
	}

	req, err := shared.CreateRequest("GET", "/v1/query/t2a_async_query_v2?task_id="+taskID, nil)
	if err != nil {
		return common.WriteError(cmd, "request_error", err.Error())
	}

	resp, err := shared.DoRequest(req)
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
		TaskID   int64  `json:"task_id"`
		Status   string `json:"status"`
		FileID   int64  `json:"file_id"`
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

	return common.WriteSuccess(cmd, map[string]any{
		"success": true,
		"task_id": apiResp.TaskID,
		"status":  apiResp.Status,
		"file_id": apiResp.FileID,
	})
}

type downloadFlags struct {
	output string
}

func newDownloadCmd() *cobra.Command {
	flags := &downloadFlags{}

	cmd := &cobra.Command{
		Use:           "download <file_id>",
		Short:         "Download async TTS result",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDownload(cmd, args[0], flags)
		},
	}

	cmd.Flags().StringVarP(&flags.output, "output", "o", "", "Output file path")
	return cmd
}

func runDownload(cmd *cobra.Command, fileID string, flags *downloadFlags) error {
	if strings.TrimSpace(fileID) == "" {
		return common.WriteError(cmd, "missing_file_id", "file_id is required")
	}
	if flags.output == "" {
		return common.WriteError(cmd, "missing_output", "output file path is required (-o)")
	}

	apiKey := shared.GetMinimaxAPIKey()
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("MINIMAX_API_KEY"))
	}

	req, err := shared.CreateRequest("GET", "/v1/files/retrieve?file_id="+fileID, nil)
	if err != nil {
		return common.WriteError(cmd, "request_error", err.Error())
	}

	resp, err := shared.DoRequest(req)
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
		File struct {
			DownloadURL string `json:"download_url"`
		} `json:"file"`
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
	if apiResp.File.DownloadURL == "" {
		return common.WriteError(cmd, "download_error", "download_url is empty")
	}

	client := &http.Client{Timeout: 5 * time.Minute}
	downloadResp, err := client.Get(apiResp.File.DownloadURL)
	if err != nil {
		return common.WriteError(cmd, "download_error", fmt.Sprintf("cannot download file: %s", err.Error()))
	}
	defer downloadResp.Body.Close()

	if downloadResp.StatusCode != http.StatusOK {
		return common.WriteError(cmd, "download_error", fmt.Sprintf("download failed with status: %s", downloadResp.Status))
	}

	dir := filepath.Dir(flags.output)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return common.WriteError(cmd, "output_error", fmt.Sprintf("failed to create directory: %s", err.Error()))
		}
	}

	outFile, err := os.Create(flags.output)
	if err != nil {
		return common.WriteError(cmd, "output_error", fmt.Sprintf("failed to create output file: %s", err.Error()))
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, downloadResp.Body); err != nil {
		return common.WriteError(cmd, "output_error", fmt.Sprintf("failed to write output file: %s", err.Error()))
	}

	absPath, err := filepath.Abs(flags.output)
	if err != nil {
		absPath = flags.output
	}

	return common.WriteSuccess(cmd, map[string]any{
		"success": true,
		"file_id": parseFileID(fileID),
		"file":    absPath,
	})
}

func parseFileID(input string) int64 {
	id, err := strconv.ParseInt(strings.TrimSpace(input), 10, 64)
	if err != nil {
		return 0
	}
	return id
}

func getTextOptional(args []string, filePath string, stdin io.Reader) (string, error) {
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
		return text, nil
	}

	if stdin != nil {
		if f, ok := stdin.(*os.File); ok {
			stat, _ := f.Stat()
			if (stat.Mode() & os.ModeCharDevice) != 0 {
				return "", nil
			}
		}
		data, err := io.ReadAll(stdin)
		if err != nil {
			return "", fmt.Errorf("cannot read stdin: %s", err.Error())
		}
		text := strings.TrimSpace(string(data))
		return text, nil
	}

	return "", nil
}
