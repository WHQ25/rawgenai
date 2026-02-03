package image

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/cli/kling/video"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

type listFlags struct {
	limit int
	page  int
}

func newListCmd() *cobra.Command {
	flags := &listFlags{}

	cmd := &cobra.Command{
		Use:           "list",
		Short:         "List image generation tasks",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd, flags)
		},
	}

	cmd.Flags().IntVarP(&flags.limit, "limit", "l", 30, "Maximum tasks to return (1-500)")
	cmd.Flags().IntVarP(&flags.page, "page", "p", 1, "Page number")

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
	token, err := video.GenerateJWT(accessKey, secretKey)
	if err != nil {
		return common.WriteError(cmd, "auth_error", fmt.Sprintf("failed to generate JWT: %s", err.Error()))
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/v1/images/generations?pageNum=%d&pageSize=%d", video.GetKlingAPIBase(), flags.page, flags.limit)
	req, err := http.NewRequest("GET", url, nil)
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
		Data    []struct {
			TaskID        string `json:"task_id"`
			TaskStatus    string `json:"task_status"`
			TaskStatusMsg string `json:"task_status_msg"`
			CreatedAt     int64  `json:"created_at"`
			UpdatedAt     int64  `json:"updated_at"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot parse response: %s", err.Error()))
	}

	if result.Code != 0 {
		return video.HandleKlingError(cmd, result.Code, result.Message)
	}

	tasks := []map[string]any{}
	for _, t := range result.Data {
		tasks = append(tasks, map[string]any{
			"task_id":    t.TaskID,
			"status":     t.TaskStatus,
			"created_at": t.CreatedAt,
			"updated_at": t.UpdatedAt,
		})
	}

	return common.WriteSuccess(cmd, map[string]any{
		"success": true,
		"tasks":   tasks,
		"count":   len(tasks),
	})
}
