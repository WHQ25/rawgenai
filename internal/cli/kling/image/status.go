package image

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/cli/kling/video"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

type statusFlags struct {
	verbose bool
}

func newStatusCmd() *cobra.Command {
	flags := &statusFlags{}

	cmd := &cobra.Command{
		Use:           "status <task_id>",
		Short:         "Get image generation status",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd, args, flags)
		},
	}

	cmd.Flags().BoolVarP(&flags.verbose, "verbose", "v", false, "Show image URLs")

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
	token, err := video.GenerateJWT(accessKey, secretKey)
	if err != nil {
		return common.WriteError(cmd, "auth_error", fmt.Sprintf("failed to generate JWT: %s", err.Error()))
	}

	// Create HTTP request
	req, err := http.NewRequest("GET", video.GetKlingAPIBase()+"/v1/images/generations/"+taskID, nil)
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot create request: %s", err.Error()))
	}

	req.Header.Set("Authorization", "Bearer "+token)

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return video.HandleAPIError(cmd, err)
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
				Images []struct {
					Index        int    `json:"index"`
					URL          string `json:"url"`
					WatermarkURL string `json:"watermark_url"`
				} `json:"images"`
			} `json:"task_result"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot parse response: %s", err.Error()))
	}

	if result.Code != 0 {
		return video.HandleKlingError(cmd, result.Code, result.Message)
	}

	if result.Data == nil {
		return common.WriteError(cmd, "response_error", "no data in response")
	}

	if result.Data.TaskStatus == "failed" {
		msg := result.Data.TaskStatusMsg
		if msg == "" {
			msg = "image generation failed"
		}
		return common.WriteError(cmd, "image_failed", msg)
	}

	output := map[string]any{
		"success": true,
		"task_id": result.Data.TaskID,
		"status":  result.Data.TaskStatus,
	}

	if result.Data.TaskResult != nil && len(result.Data.TaskResult.Images) > 0 {
		output["image_count"] = len(result.Data.TaskResult.Images)
		if flags.verbose {
			images := make([]map[string]any, 0, len(result.Data.TaskResult.Images))
			for _, img := range result.Data.TaskResult.Images {
				item := map[string]any{
					"index": img.Index,
					"url":   img.URL,
				}
				if img.WatermarkURL != "" {
					item["watermark_url"] = img.WatermarkURL
				}
				images = append(images, item)
			}
			output["images"] = images
		}
	}

	return common.WriteSuccess(cmd, output)
}
