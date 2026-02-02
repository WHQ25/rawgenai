package video

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

const videosPath = "/videos"

type statusResponse struct {
	Success   bool    `json:"success"`
	RequestID string  `json:"request_id"`
	Status    string  `json:"status"`
	Progress  float64 `json:"progress,omitempty"`
	Error     string  `json:"error_message,omitempty"`
}

// API response type
type xaiVideoStatusResponse struct {
	RequestID string         `json:"request_id"`
	Status    string         `json:"status"`
	Progress  float64        `json:"progress,omitempty"`
	VideoURL  string         `json:"video_url,omitempty"`
	Error     *xaiVideoError `json:"error,omitempty"`
}

var statusCmd = newStatusCmd()

func newStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "status <request_id>",
		Short:         "Get video generation status",
		Long:          "Query the status of a video generation or edit operation.",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd, args)
		},
	}

	return cmd
}

func runStatus(cmd *cobra.Command, args []string) error {
	requestID := strings.TrimSpace(args[0])
	if requestID == "" {
		return common.WriteError(cmd, "missing_request_id", "request_id is required")
	}

	// Check API key
	apiKey := config.GetAPIKey("XAI_API_KEY")
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("XAI_API_KEY"))
	}

	// Make request
	url := fmt.Sprintf("%s%s/%s", xaiBaseURL, videosPath, requestID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return common.WriteError(cmd, "request_error", err.Error())
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return handleHTTPError(cmd, err)
	}
	defer resp.Body.Close()

	// Parse response
	var apiResp xaiVideoStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot parse response: %s", err.Error()))
	}

	// Check for API error
	if apiResp.Error != nil {
		return handleAPIError(cmd, resp.StatusCode, apiResp.Error)
	}

	if resp.StatusCode != http.StatusOK {
		return common.WriteError(cmd, "api_error", fmt.Sprintf("API returned status %d", resp.StatusCode))
	}

	result := statusResponse{
		Success:   true,
		RequestID: apiResp.RequestID,
		Status:    apiResp.Status,
		Progress:  apiResp.Progress,
	}

	// Check if there was an error in the video generation
	if apiResp.Status == "failed" && apiResp.Error != nil {
		result.Error = apiResp.Error.Message
	}

	return common.WriteSuccess(cmd, result)
}
