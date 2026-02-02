package video

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

var validT2VModels = map[string]bool{
	"kling-v1":           true,
	"kling-v1-6":         true,
	"kling-v2-master":    true,
	"kling-v2-1-master":  true,
	"kling-v2-5-turbo":   true,
	"kling-v2-6":         true,
}

var validT2VDurations = map[string]bool{"5": true, "10": true}

type text2videoFlags struct {
	negativePrompt string
	model          string
	mode           string
	duration       string
	ratio          string
	cfgScale       float64
	cameraControl  string
	sound          bool
	watermark      bool
	promptFile     string
}

func newText2VideoCmd() *cobra.Command {
	flags := &text2videoFlags{}

	cmd := &cobra.Command{
		Use:           "create-from-text [prompt]",
		Short:         "Create video from text (legacy models)",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runText2Video(cmd, args, flags)
		},
	}

	cmd.Flags().StringVar(&flags.negativePrompt, "negative", "", "Negative prompt")
	cmd.Flags().StringVarP(&flags.model, "model", "m", "kling-v1", "Model: kling-v1, kling-v1-6, kling-v2-master, kling-v2-1-master, kling-v2-5-turbo, kling-v2-6")
	cmd.Flags().StringVar(&flags.mode, "mode", "std", "Generation mode: std, pro")
	cmd.Flags().StringVarP(&flags.duration, "duration", "d", "5", "Video duration: 5, 10")
	cmd.Flags().StringVarP(&flags.ratio, "ratio", "r", "16:9", "Aspect ratio: 16:9, 9:16, 1:1")
	cmd.Flags().Float64Var(&flags.cfgScale, "cfg-scale", 0.5, "Prompt adherence (0-1), not supported by v2.x models")
	cmd.Flags().StringVar(&flags.cameraControl, "camera-control", "", "Camera control JSON (type, config)")
	cmd.Flags().BoolVar(&flags.sound, "sound", false, "Generate sound (v2.6+ only)")
	cmd.Flags().BoolVar(&flags.watermark, "watermark", false, "Include watermark")
	cmd.Flags().StringVarP(&flags.promptFile, "prompt-file", "f", "", "Read prompt from file")

	return cmd
}

func runText2Video(cmd *cobra.Command, args []string, flags *text2videoFlags) error {
	// Get prompt
	prompt, err := getPrompt(args, flags.promptFile, cmd.InOrStdin())
	if err != nil {
		return common.WriteError(cmd, "missing_prompt", err.Error())
	}

	// Validate model
	if !validT2VModels[flags.model] {
		return common.WriteError(cmd, "invalid_model", fmt.Sprintf("invalid model '%s'", flags.model))
	}

	// Validate mode
	if !validModes[flags.mode] {
		return common.WriteError(cmd, "invalid_mode", fmt.Sprintf("invalid mode '%s', use std or pro", flags.mode))
	}

	// Validate duration
	if !validT2VDurations[flags.duration] {
		return common.WriteError(cmd, "invalid_duration", "duration must be 5 or 10")
	}

	// Validate ratio
	if !validRatios[flags.ratio] {
		return common.WriteError(cmd, "invalid_ratio", fmt.Sprintf("invalid ratio '%s', use 16:9, 9:16, or 1:1", flags.ratio))
	}

	// Validate cfg_scale
	if flags.cfgScale < 0 || flags.cfgScale > 1 {
		return common.WriteError(cmd, "invalid_cfg_scale", "cfg-scale must be between 0 and 1")
	}

	// Validate camera control compatibility
	// kling-v1: only supports std 5s
	if flags.cameraControl != "" {
		if flags.model != "kling-v1" {
			return common.WriteError(cmd, "incompatible_camera_control", "camera control only supported by kling-v1")
		}
		if flags.mode != "std" || flags.duration != "5" {
			return common.WriteError(cmd, "incompatible_camera_control", "camera control only supported in std mode with 5s duration")
		}
	}

	// Validate sound compatibility (only v2.6)
	if flags.sound && flags.model != "kling-v2-6" {
		return common.WriteError(cmd, "incompatible_sound", "sound generation only supported by kling-v2-6")
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
		"model_name":   flags.model,
		"prompt":       prompt,
		"mode":         flags.mode,
		"duration":     flags.duration,
		"aspect_ratio": flags.ratio,
	}

	if flags.negativePrompt != "" {
		body["negative_prompt"] = flags.negativePrompt
	}

	// cfg_scale not supported by v2.x models
	if !isV2Model(flags.model) {
		body["cfg_scale"] = flags.cfgScale
	}

	// Parse and add camera control if provided
	if flags.cameraControl != "" {
		var cameraControl map[string]any
		if err := json.Unmarshal([]byte(flags.cameraControl), &cameraControl); err != nil {
			return common.WriteError(cmd, "invalid_camera_control", fmt.Sprintf("invalid camera control JSON: %s", err.Error()))
		}
		body["camera_control"] = cameraControl
	}

	if flags.sound {
		body["sound"] = "on"
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
	req, err := http.NewRequest("POST", getKlingAPIBase()+"/v1/videos/text2video", bytes.NewReader(jsonBody))
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

	if result.Code != 0 {
		return handleKlingError(cmd, result.Code, result.Message)
	}

	if result.Data == nil {
		return common.WriteError(cmd, "response_error", "no data in response")
	}

	return common.WriteSuccess(cmd, map[string]any{
		"success": true,
		"task_id": result.Data.TaskID,
		"status":  result.Data.TaskStatus,
	})
}

func isV2Model(model string) bool {
	return model == "kling-v2-master" || model == "kling-v2-1-master" ||
		model == "kling-v2-5-turbo" || model == "kling-v2-6"
}
