package video

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/spf13/cobra"
)

const videoEditsPath = "/videos/edits"

type editFlags struct {
	video      string
	promptFile string
}

type editResponse struct {
	Success   bool   `json:"success"`
	RequestID string `json:"request_id"`
	Status    string `json:"status"`
}

// API response type
type xaiVideoEditResponse struct {
	RequestID string         `json:"request_id"`
	Status    string         `json:"status"`
	Error     *xaiVideoError `json:"error,omitempty"`
}

var editCmd = newEditCmd()

func newEditCmd() *cobra.Command {
	flags := &editFlags{}

	cmd := &cobra.Command{
		Use:   "edit [prompt]",
		Short: "Edit a video",
		Long: `Edit a video using xAI Grok API.

Requires a video URL (--video) and a prompt describing the desired changes.

IMPORTANT: Save the request_id from the response. You need it to check status
and download the edited video.`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEdit(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.video, "video", "v", "", "Input video URL (required)")
	cmd.Flags().StringVar(&flags.promptFile, "prompt-file", "", "Read prompt from file")

	return cmd
}

func runEdit(cmd *cobra.Command, args []string, flags *editFlags) error {
	// Get prompt
	prompt, err := getPrompt(args, flags.promptFile, cmd.InOrStdin())
	if err != nil {
		return common.WriteError(cmd, "missing_prompt", err.Error())
	}

	// Validate video URL
	if flags.video == "" {
		return common.WriteError(cmd, "missing_video", "video URL is required, use --video flag")
	}

	// Check API key
	apiKey := os.Getenv("XAI_API_KEY")
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", "XAI_API_KEY environment variable is not set")
	}

	// Build request body
	reqBody := map[string]any{
		"model":  "grok-2-video",
		"prompt": prompt,
		"video":  flags.video,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return common.WriteError(cmd, "json_error", err.Error())
	}

	// Make request
	req, err := http.NewRequest("POST", xaiBaseURL+videoEditsPath, bytes.NewReader(jsonBody))
	if err != nil {
		return common.WriteError(cmd, "request_error", err.Error())
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return handleHTTPError(cmd, err)
	}
	defer resp.Body.Close()

	// Parse response
	var apiResp xaiVideoEditResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot parse response: %s", err.Error()))
	}

	// Check for API error
	if apiResp.Error != nil {
		return handleAPIError(cmd, resp.StatusCode, apiResp.Error)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return common.WriteError(cmd, "api_error", fmt.Sprintf("API returned status %d", resp.StatusCode))
	}

	status := apiResp.Status
	if status == "" {
		status = "pending"
	}

	return common.WriteSuccess(cmd, editResponse{
		Success:   true,
		RequestID: apiResp.RequestID,
		Status:    status,
	})
}
