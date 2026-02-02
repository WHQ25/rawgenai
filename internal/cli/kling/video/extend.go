package video

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

type extendFlags struct {
	prompt         string
	negativePrompt string
	cfgScale       float64
	watermark      bool
}

func newExtendCmd() *cobra.Command {
	flags := &extendFlags{}

	cmd := &cobra.Command{
		Use:           "extend <video_id>",
		Short:         "Extend an existing video",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExtend(cmd, args, flags)
		},
	}

	cmd.Flags().StringVar(&flags.prompt, "prompt", "", "Prompt for the extended portion")
	cmd.Flags().StringVar(&flags.negativePrompt, "negative", "", "Negative prompt")
	cmd.Flags().Float64Var(&flags.cfgScale, "cfg-scale", 0.5, "Prompt adherence (0-1)")
	cmd.Flags().BoolVar(&flags.watermark, "watermark", false, "Include watermark")

	return cmd
}

func runExtend(cmd *cobra.Command, args []string, flags *extendFlags) error {
	// Validate video ID
	if len(args) == 0 || strings.TrimSpace(args[0]) == "" {
		return common.WriteError(cmd, "missing_video_id", "video ID is required")
	}
	videoID := args[0]

	// Validate cfg-scale
	if flags.cfgScale < 0 || flags.cfgScale > 1 {
		return common.WriteError(cmd, "invalid_cfg_scale", "cfg-scale must be between 0 and 1")
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

	// Build request body
	body := map[string]any{
		"video_id":  videoID,
		"cfg_scale": flags.cfgScale,
	}

	if flags.prompt != "" {
		body["prompt"] = flags.prompt
	}

	if flags.negativePrompt != "" {
		body["negative_prompt"] = flags.negativePrompt
	}

	if flags.watermark {
		body["watermark_info"] = map[string]bool{"enabled": true}
	}

	// Serialize request
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot serialize request: %s", err.Error()))
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", getKlingAPIBase()+"/v1/videos/video-extend", bytes.NewReader(jsonBody))
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot create request: %s", err.Error()))
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Send request
	client := &http.Client{Timeout: 60 * time.Second}
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
			TaskID     string `json:"task_id"`
			TaskStatus string `json:"task_status"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot parse response: %s", err.Error()))
	}

	// Check for errors
	if result.Code != 0 {
		return handleKlingError(cmd, result.Code, result.Message)
	}

	if result.Data == nil {
		return common.WriteError(cmd, "response_error", "no data in response")
	}

	// Return success
	return common.WriteSuccess(cmd, map[string]any{
		"success": true,
		"task_id": result.Data.TaskID,
		"status":  result.Data.TaskStatus,
	})
}
