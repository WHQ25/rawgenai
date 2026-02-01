package video

import (
	"errors"
	"fmt"
	"strings"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	oai "github.com/openai/openai-go/v3"
	"github.com/spf13/cobra"
)

// Cmd is the video parent command
var Cmd = &cobra.Command{
	Use:   "video",
	Short: "Video generation commands using OpenAI Sora",
	Long:  "Commands for video generation using OpenAI Sora models (sora-2, sora-2-pro).",
}

func init() {
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(statusCmd)
	Cmd.AddCommand(downloadCmd)
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(deleteCmd)
	Cmd.AddCommand(remixCmd)
}

func handleAPIError(cmd *cobra.Command, err error) error {
	var apiErr *oai.Error
	if errors.As(err, &apiErr) {
		switch apiErr.StatusCode {
		case 400:
			if strings.Contains(strings.ToLower(apiErr.Message), "content") || strings.Contains(strings.ToLower(apiErr.Message), "policy") {
				return common.WriteError(cmd, "content_policy", apiErr.Message)
			}
			if strings.Contains(strings.ToLower(apiErr.Message), "model") {
				return common.WriteError(cmd, "invalid_model", apiErr.Message)
			}
			return common.WriteError(cmd, "invalid_request", apiErr.Message)
		case 401:
			return common.WriteError(cmd, "invalid_api_key", "API key is invalid or revoked")
		case 403:
			return common.WriteError(cmd, "region_not_supported", "Region/country not supported")
		case 404:
			return common.WriteError(cmd, "video_not_found", "Video not found")
		case 429:
			if strings.Contains(apiErr.Message, "quota") {
				return common.WriteError(cmd, "quota_exceeded", apiErr.Message)
			}
			return common.WriteError(cmd, "rate_limit", apiErr.Message)
		case 500:
			return common.WriteError(cmd, "server_error", "OpenAI server error")
		case 503:
			return common.WriteError(cmd, "server_overloaded", "OpenAI server overloaded")
		default:
			return common.WriteError(cmd, "api_error", apiErr.Message)
		}
	}

	if strings.Contains(err.Error(), "timeout") {
		return common.WriteError(cmd, "timeout", "Request timed out")
	}
	return common.WriteError(cmd, "connection_error", fmt.Sprintf("Cannot connect to OpenAI API: %s", err.Error()))
}
