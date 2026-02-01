package video

import (
	"strings"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/spf13/cobra"
)

const xaiBaseURL = "https://api.x.ai/v1"

// Cmd is the video parent command
var Cmd = &cobra.Command{
	Use:   "video",
	Short: "Video generation commands using xAI Grok",
	Long: `Commands for video generation and editing using xAI Grok API.

IMPORTANT: Video operations are asynchronous. Save the request_id returned by
'create' and 'edit' commands. You need it to check status and download videos.`,
}

func init() {
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(editCmd)
	Cmd.AddCommand(statusCmd)
	Cmd.AddCommand(downloadCmd)
}

// xaiVideoError represents an error from the xAI API
type xaiVideoError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

func handleAPIError(cmd *cobra.Command, statusCode int, xaiErr *xaiVideoError) error {
	switch statusCode {
	case 400:
		return common.WriteError(cmd, "invalid_request", xaiErr.Message)
	case 401:
		return common.WriteError(cmd, "invalid_api_key", "API key is invalid or revoked")
	case 403:
		return common.WriteError(cmd, "permission_denied", "API key lacks required permissions")
	case 404:
		return common.WriteError(cmd, "not_found", "request ID does not exist")
	case 429:
		if strings.Contains(xaiErr.Message, "quota") {
			return common.WriteError(cmd, "quota_exceeded", xaiErr.Message)
		}
		return common.WriteError(cmd, "rate_limit", xaiErr.Message)
	case 500:
		return common.WriteError(cmd, "server_error", "xAI server error")
	case 503:
		return common.WriteError(cmd, "server_overloaded", "xAI server overloaded")
	default:
		return common.WriteError(cmd, "api_error", xaiErr.Message)
	}
}

func handleHTTPError(cmd *cobra.Command, err error) error {
	errStr := err.Error()
	if strings.Contains(errStr, "timeout") {
		return common.WriteError(cmd, "timeout", "request timed out")
	}
	if strings.Contains(errStr, "connection") {
		return common.WriteError(cmd, "connection_error", "cannot connect to xAI API")
	}
	return common.WriteError(cmd, "network_error", err.Error())
}
