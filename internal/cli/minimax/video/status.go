package video

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/cli/minimax/shared"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

type statusResponse struct {
	Success     bool   `json:"success"`
	TaskID      string `json:"task_id,omitempty"`
	Status      string `json:"status,omitempty"`
	FileID      string `json:"file_id,omitempty"`
	VideoWidth  int    `json:"video_width,omitempty"`
	VideoHeight int    `json:"video_height,omitempty"`
}

func newStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "status <task_id>",
		Short:         "Query video generation task status",
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
	if taskID == "" {
		return common.WriteError(cmd, "missing_task_id", "task_id is required")
	}

	apiKey := shared.GetMinimaxAPIKey()
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("MINIMAX_API_KEY"))
	}

	req, err := shared.CreateRequest("GET", "/v1/query/video_generation?task_id="+taskID, nil)
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
		TaskID      string `json:"task_id"`
		Status      string `json:"status"`
		FileID      string `json:"file_id"`
		VideoWidth  int    `json:"video_width"`
		VideoHeight int    `json:"video_height"`
		BaseResp    struct {
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

	return common.WriteSuccess(cmd, statusResponse{
		Success:     true,
		TaskID:      apiResp.TaskID,
		Status:      apiResp.Status,
		FileID:      apiResp.FileID,
		VideoWidth:  apiResp.VideoWidth,
		VideoHeight: apiResp.VideoHeight,
	})
}
