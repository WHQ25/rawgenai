package video

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

type listFlags struct {
	limit    int
	page     int
	taskType string
}

func newListCmd() *cobra.Command {
	flags := &listFlags{}

	cmd := &cobra.Command{
		Use:           "list",
		Short:         "List video generation tasks",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd, flags)
		},
	}

	cmd.Flags().IntVarP(&flags.limit, "limit", "l", 30, "Maximum tasks to return (1-500)")
	cmd.Flags().IntVarP(&flags.page, "page", "p", 1, "Page number")
	cmd.Flags().StringVarP(&flags.taskType, "type", "t", "create", "Task type: create, text2video, image2video, motion-control, avatar, extend, add-sound")

	return cmd
}

func runList(cmd *cobra.Command, flags *listFlags) error {
	// Validate limit
	if flags.limit < 1 || flags.limit > 500 {
		return common.WriteError(cmd, "invalid_limit", "limit must be between 1 and 500")
	}

	// Validate page
	if flags.page < 1 {
		return common.WriteError(cmd, "invalid_page", "page must be at least 1")
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

	// Determine endpoint based on task type
	endpoint := "/v1/videos/omni-video"
	switch flags.taskType {
	case "text2video":
		endpoint = "/v1/videos/text2video"
	case "image2video":
		endpoint = "/v1/videos/image2video"
	case "extend":
		endpoint = "/v1/videos/video-extend"
	case "add-sound":
		endpoint = "/v1/audio/video-to-audio"
	case "motion-control":
		endpoint = "/v1/videos/motion-control"
	case "avatar":
		endpoint = "/v1/videos/avatar/image2video"
	}

	// Create HTTP request
	url := fmt.Sprintf("%s%s?pageNum=%d&pageSize=%d", getKlingAPIBase(), endpoint, flags.page, flags.limit)
	req, err := http.NewRequest("GET", url, nil)
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
			Tasks []struct {
				TaskID        string `json:"task_id"`
				TaskStatus    string `json:"task_status"`
				TaskStatusMsg string `json:"task_status_msg"`
				CreatedAt     int64  `json:"created_at"`
			} `json:"task_list"`
			Total int `json:"total"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot parse response: %s", err.Error()))
	}

	// Check for errors
	if result.Code != 0 {
		return handleKlingError(cmd, result.Code, result.Message)
	}

	// Build response
	tasks := []map[string]any{}
	if result.Data != nil && result.Data.Tasks != nil {
		for _, t := range result.Data.Tasks {
			tasks = append(tasks, map[string]any{
				"task_id":    t.TaskID,
				"status":     t.TaskStatus,
				"created_at": t.CreatedAt,
			})
		}
	}

	count := 0
	if result.Data != nil {
		count = result.Data.Total
	}

	return common.WriteSuccess(cmd, map[string]any{
		"success": true,
		"tasks":   tasks,
		"count":   count,
	})
}
