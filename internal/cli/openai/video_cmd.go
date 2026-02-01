package openai

import (
	"errors"
	"fmt"
	"strings"

	oai "github.com/openai/openai-go/v3"
	"github.com/spf13/cobra"
)

var videoCmd = &cobra.Command{
	Use:   "video",
	Short: "Video generation commands using OpenAI Sora",
	Long:  "Commands for video generation using OpenAI Sora models (sora-2, sora-2-pro).",
}

func init() {
	videoCmd.AddCommand(videoCreateCmd)
	videoCmd.AddCommand(videoStatusCmd)
	videoCmd.AddCommand(videoDownloadCmd)
	videoCmd.AddCommand(videoListCmd)
	videoCmd.AddCommand(videoDeleteCmd)
	videoCmd.AddCommand(videoRemixCmd)
}

func handleVideoAPIError(cmd *cobra.Command, err error) error {
	var apiErr *oai.Error
	if errors.As(err, &apiErr) {
		switch apiErr.StatusCode {
		case 400:
			if strings.Contains(strings.ToLower(apiErr.Message), "content") || strings.Contains(strings.ToLower(apiErr.Message), "policy") {
				return writeError(cmd, "content_policy", apiErr.Message)
			}
			if strings.Contains(strings.ToLower(apiErr.Message), "model") {
				return writeError(cmd, "invalid_model", apiErr.Message)
			}
			return writeError(cmd, "invalid_request", apiErr.Message)
		case 401:
			return writeError(cmd, "invalid_api_key", "API key is invalid or revoked")
		case 403:
			return writeError(cmd, "region_not_supported", "Region/country not supported")
		case 404:
			return writeError(cmd, "video_not_found", "Video not found")
		case 429:
			if strings.Contains(apiErr.Message, "quota") {
				return writeError(cmd, "quota_exceeded", apiErr.Message)
			}
			return writeError(cmd, "rate_limit", apiErr.Message)
		case 500:
			return writeError(cmd, "server_error", "OpenAI server error")
		case 503:
			return writeError(cmd, "server_overloaded", "OpenAI server overloaded")
		default:
			return writeError(cmd, "api_error", apiErr.Message)
		}
	}

	if strings.Contains(err.Error(), "timeout") {
		return writeError(cmd, "timeout", "Request timed out")
	}
	return writeError(cmd, "connection_error", fmt.Sprintf("Cannot connect to OpenAI API: %s", err.Error()))
}
