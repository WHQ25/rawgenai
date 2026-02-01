package video

import (
	"strings"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/spf13/cobra"
)

// Cmd is the video parent command
var Cmd = &cobra.Command{
	Use:   "video",
	Short: "Video generation commands using Google Veo",
	Long: `Commands for video generation using Google Veo 3.1 models (veo-3.1, veo-3.1-fast).

IMPORTANT: Save the operation_id returned by 'create' and 'extend' commands.
You need it to check status, download video, and extend. Videos are stored
on Google's servers for only 2 days before automatic deletion.`,
}

func init() {
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(extendCmd)
	Cmd.AddCommand(statusCmd)
	Cmd.AddCommand(downloadCmd)
}

func handleAPIError(cmd *cobra.Command, err error) error {
	errStr := err.Error()

	// Check for common error patterns
	if strings.Contains(errStr, "401") || (strings.Contains(errStr, "invalid") && strings.Contains(errStr, "key")) {
		return common.WriteError(cmd, "invalid_api_key", "API key is invalid or revoked")
	}
	if strings.Contains(errStr, "403") || strings.Contains(errStr, "permission") {
		return common.WriteError(cmd, "permission_denied", "API key lacks required permissions")
	}
	if strings.Contains(errStr, "404") {
		return common.WriteError(cmd, "operation_not_found", "Operation ID does not exist")
	}
	if strings.Contains(errStr, "429") {
		if strings.Contains(errStr, "quota") {
			return common.WriteError(cmd, "quota_exceeded", "API quota exhausted")
		}
		return common.WriteError(cmd, "rate_limit", "too many requests")
	}
	if strings.Contains(errStr, "safety") || strings.Contains(errStr, "policy") {
		return common.WriteError(cmd, "content_policy", "content violates safety policy")
	}
	if strings.Contains(errStr, "timeout") {
		return common.WriteError(cmd, "timeout", "request timed out")
	}
	if strings.Contains(errStr, "connection") {
		return common.WriteError(cmd, "connection_error", "cannot connect to Gemini API")
	}
	if strings.Contains(errStr, "500") {
		return common.WriteError(cmd, "server_error", "Gemini server error")
	}
	if strings.Contains(errStr, "503") {
		return common.WriteError(cmd, "server_overloaded", "Gemini server overloaded")
	}

	return common.WriteError(cmd, "api_error", err.Error())
}
