package video

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

type statusFlags struct {
	taskType string
	verbose  bool
}

func newStatusCmd() *cobra.Command {
	flags := &statusFlags{}

	cmd := &cobra.Command{
		Use:           "status <task_id>",
		Short:         "Get video generation status",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.taskType, "type", "t", "create", "Task type: create, text2video, image2video, motion-control, avatar, extend, add-sound")
	cmd.Flags().BoolVarP(&flags.verbose, "verbose", "v", false, "Show full output including URLs")

	return cmd
}

func runStatus(cmd *cobra.Command, args []string, flags *statusFlags) error {
	// Validate task ID
	if len(args) == 0 || strings.TrimSpace(args[0]) == "" {
		return common.WriteError(cmd, "missing_task_id", "task ID is required")
	}
	taskID := args[0]

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

	// Determine endpoint based on task type
	endpoint := "/v1/videos/omni-video/"
	switch flags.taskType {
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
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot create request: %s", err.Error()))
	}

	req.Header.Set("Authorization", "Bearer "+token)

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return handleAPIError(cmd, err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot read response: %s", err.Error()))
	}

	// Parse response
	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    *struct {
			TaskID        string `json:"task_id"`
			TaskStatus    string `json:"task_status"`
			TaskStatusMsg string `json:"task_status_msg"`
			TaskResult    *struct {
				Videos []struct {
					ID           string `json:"id"`
					URL          string `json:"url"`
					WatermarkURL string `json:"watermark_url"`
					Duration     string `json:"duration"`
				} `json:"videos"`
				Audios []struct {
					ID          string `json:"id"`
					URLMP3      string `json:"url_mp3"`
					URLWAV      string `json:"url_wav"`
					DurationMP3 string `json:"duration_mp3"`
					DurationWAV string `json:"duration_wav"`
				} `json:"audios"`
			} `json:"task_result"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot parse response: %s", err.Error()))
	}

	// Check for errors
	if result.Code != 0 {
		return handleKlingError(cmd, result.Code, result.Message)
	}

	if result.Data == nil {
		return common.WriteError(cmd, "response_error", "no data in response")
	}

	// Handle failed status
	if result.Data.TaskStatus == "failed" {
		msg := result.Data.TaskStatusMsg
		if msg == "" {
			msg = "video generation failed"
		}
		return common.WriteError(cmd, "video_failed", msg)
	}

	// Build response
	output := map[string]any{
		"success": true,
		"task_id": result.Data.TaskID,
		"status":  result.Data.TaskStatus,
	}

	if result.Data.TaskStatus == "succeed" && result.Data.TaskResult != nil {
		if len(result.Data.TaskResult.Videos) > 0 {
			video := result.Data.TaskResult.Videos[0]
			output["video_id"] = video.ID
			output["duration"] = video.Duration
			if flags.verbose {
				output["video_url"] = video.URL
				if video.WatermarkURL != "" {
					output["watermark_url"] = video.WatermarkURL
				}
			}
		}
		if len(result.Data.TaskResult.Audios) > 0 && flags.verbose {
			audio := result.Data.TaskResult.Audios[0]
			output["audio_mp3_url"] = audio.URLMP3
			output["audio_wav_url"] = audio.URLWAV
		}
	}

	return common.WriteSuccess(cmd, output)
}
