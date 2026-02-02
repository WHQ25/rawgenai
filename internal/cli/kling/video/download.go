package video

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

type downloadFlags struct {
	output    string
	taskType  string
	watermark bool
	format    string
}

func newDownloadCmd() *cobra.Command {
	flags := &downloadFlags{}

	cmd := &cobra.Command{
		Use:           "download <task_id>",
		Short:         "Download completed video",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDownload(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.output, "output", "o", "", "Output file path")
	cmd.Flags().StringVarP(&flags.taskType, "type", "t", "create", "Task type: create, text2video, image2video, motion-control, avatar, extend, add-sound")
	cmd.Flags().BoolVar(&flags.watermark, "watermark", false, "Download watermarked version")
	cmd.Flags().StringVar(&flags.format, "format", "video", "Download format: video, mp3, wav (mp3/wav only for add-sound)")

	return cmd
}

func runDownload(cmd *cobra.Command, args []string, flags *downloadFlags) error {
	// Validate task ID
	if len(args) == 0 || strings.TrimSpace(args[0]) == "" {
		return common.WriteError(cmd, "missing_task_id", "task ID is required")
	}
	taskID := args[0]

	// Validate output
	if flags.output == "" {
		return common.WriteError(cmd, "missing_output", "output file path is required (-o)")
	}

	// Validate format flag
	validFormats := map[string]bool{"video": true, "mp3": true, "wav": true}
	if !validFormats[flags.format] {
		return common.WriteError(cmd, "invalid_format", "format must be video, mp3, or wav")
	}

	// Audio formats only allowed for add-sound
	if (flags.format == "mp3" || flags.format == "wav") && flags.taskType != "add-sound" {
		return common.WriteError(cmd, "invalid_format", "mp3/wav format only supported for --type add-sound")
	}

	// Validate output file extension
	lowerOutput := strings.ToLower(flags.output)
	expectedExt := ".mp4"
	if flags.format == "mp3" {
		expectedExt = ".mp3"
	} else if flags.format == "wav" {
		expectedExt = ".wav"
	}
	if !strings.HasSuffix(lowerOutput, expectedExt) {
		return common.WriteError(cmd, "invalid_format", fmt.Sprintf("output file must be %s for format %s", expectedExt, flags.format))
	}

	// Check API keys
	accessKey := config.GetAPIKey("KLING_ACCESS_KEY")
	secretKey := config.GetAPIKey("KLING_SECRET_KEY")
	if accessKey == "" || secretKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("KLING_ACCESS_KEY")+" and "+config.GetMissingKeyMessage("KLING_SECRET_KEY"))
	}

	// Generate JWT token
	token, err := generateJWT(accessKey, secretKey)
	if err != nil {
		return common.WriteError(cmd, "auth_error", fmt.Sprintf("failed to generate JWT: %s", err.Error()))
	}

	// First, get the download URL from status
	downloadURL, err := getDownloadURL(token, taskID, flags.taskType, flags.watermark, flags.format)
	if err != nil {
		return err
	}

	// Download the file
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(downloadURL)
	if err != nil {
		return common.WriteError(cmd, "download_error", fmt.Sprintf("cannot download file: %s", err.Error()))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return common.WriteError(cmd, "download_error", fmt.Sprintf("download failed with status: %d", resp.StatusCode))
	}

	// Create output directory if needed
	dir := filepath.Dir(flags.output)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return common.WriteError(cmd, "write_error", fmt.Sprintf("cannot create directory: %s", err.Error()))
		}
	}

	// Create output file
	outFile, err := os.Create(flags.output)
	if err != nil {
		return common.WriteError(cmd, "write_error", fmt.Sprintf("cannot create file: %s", err.Error()))
	}
	defer outFile.Close()

	// Copy data
	if _, err := io.Copy(outFile, resp.Body); err != nil {
		return common.WriteError(cmd, "write_error", fmt.Sprintf("cannot write file: %s", err.Error()))
	}

	// Get absolute path
	absPath, err := filepath.Abs(flags.output)
	if err != nil {
		absPath = flags.output
	}

	return common.WriteSuccess(cmd, map[string]any{
		"success": true,
		"task_id": taskID,
		"file":    absPath,
	})
}

func getDownloadURL(token, taskID, taskType string, watermark bool, format string) (string, error) {
	// Determine endpoint based on task type
	endpoint := "/v1/videos/omni-video/"
	switch taskType {
	case "text2video":
		endpoint = "/v1/videos/text2video/"
	case "image2video":
		endpoint = "/v1/videos/image2video/"
	case "extend":
		endpoint = "/v1/videos/video-extend/"
	case "add-sound":
		endpoint = "/v1/audio/video-to-audio/"
	case "motion-control":
		endpoint = "/v1/videos/motion-control/"
	case "avatar":
		endpoint = "/v1/videos/avatar/image2video/"
	}

	// Create HTTP request
	req, err := http.NewRequest("GET", getKlingAPIBase()+endpoint+taskID, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create request: %s", err.Error())
	}

	req.Header.Set("Authorization", "Bearer "+token)

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot get status: %s", err.Error())
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("cannot read response: %s", err.Error())
	}

	// Parse response - includes audio URLs for add-sound tasks
	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    *struct {
			TaskStatus string `json:"task_status"`
			TaskResult *struct {
				Videos []struct {
					URL          string `json:"url"`
					WatermarkURL string `json:"watermark_url"`
				} `json:"videos"`
				Audios []struct {
					URLMP3 string `json:"url_mp3"`
					URLWAV string `json:"url_wav"`
				} `json:"audios"`
			} `json:"task_result"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("cannot parse response: %s", err.Error())
	}

	if result.Code != 0 {
		return "", fmt.Errorf("API error: %s", result.Message)
	}

	if result.Data == nil {
		return "", fmt.Errorf("no data in response")
	}

	if result.Data.TaskStatus != "succeed" {
		return "", fmt.Errorf("task not completed (status: %s)", result.Data.TaskStatus)
	}

	// Return audio URL if requested
	if format == "mp3" || format == "wav" {
		if result.Data.TaskResult == nil || len(result.Data.TaskResult.Audios) == 0 {
			return "", fmt.Errorf("no audio in result")
		}
		audio := result.Data.TaskResult.Audios[0]
		if format == "mp3" {
			if audio.URLMP3 == "" {
				return "", fmt.Errorf("no MP3 audio in result")
			}
			return audio.URLMP3, nil
		}
		if audio.URLWAV == "" {
			return "", fmt.Errorf("no WAV audio in result")
		}
		return audio.URLWAV, nil
	}

	// Return video URL
	if result.Data.TaskResult == nil || len(result.Data.TaskResult.Videos) == 0 {
		return "", fmt.Errorf("no video in result")
	}

	video := result.Data.TaskResult.Videos[0]
	if watermark && video.WatermarkURL != "" {
		return video.WatermarkURL, nil
	}
	return video.URL, nil
}
