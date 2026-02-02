package video

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

var validI2VModels = map[string]bool{
	"kling-v1":          true,
	"kling-v1-5":        true,
	"kling-v1-6":        true,
	"kling-v2-master":   true,
	"kling-v2-1":        true,
	"kling-v2-1-master": true,
	"kling-v2-5-turbo":  true,
	"kling-v2-6":        true,
}

type image2videoFlags struct {
	firstFrame     string
	lastFrame      string
	negativePrompt string
	model          string
	mode           string
	duration       string
	cfgScale       float64
	cameraControl  string
	staticMask     string
	dynamicMask    string
	voice          []string
	sound          bool
	watermark      bool
	promptFile     string
}

func newImage2VideoCmd() *cobra.Command {
	flags := &image2videoFlags{}

	cmd := &cobra.Command{
		Use:           "create-from-image [prompt]",
		Short:         "Create video from image (legacy models)",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runImage2Video(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.firstFrame, "first-frame", "i", "", "First frame image (required)")
	cmd.Flags().StringVar(&flags.lastFrame, "last-frame", "", "Last frame image")
	cmd.Flags().StringVar(&flags.negativePrompt, "negative", "", "Negative prompt")
	cmd.Flags().StringVarP(&flags.model, "model", "m", "kling-v1", "Model: kling-v1, kling-v1-5, kling-v1-6, kling-v2-master, kling-v2-1, kling-v2-1-master, kling-v2-5-turbo, kling-v2-6")
	cmd.Flags().StringVar(&flags.mode, "mode", "std", "Generation mode: std, pro")
	cmd.Flags().StringVarP(&flags.duration, "duration", "d", "5", "Video duration: 5, 10")
	cmd.Flags().Float64Var(&flags.cfgScale, "cfg-scale", 0.5, "Prompt adherence (0-1), not supported by v2.x models")
	cmd.Flags().StringVar(&flags.cameraControl, "camera-control", "", "Camera control JSON (type, config)")
	cmd.Flags().StringVar(&flags.staticMask, "static-mask", "", "Static brush mask image (local file or URL)")
	cmd.Flags().StringVar(&flags.dynamicMask, "dynamic-mask", "", "Dynamic mask JSON (mask image + trajectories)")
	cmd.Flags().StringSliceVar(&flags.voice, "voice", nil, "Voice ID(s) for v2.6+ models")
	cmd.Flags().BoolVar(&flags.sound, "sound", false, "Generate sound (v2.6+ only)")
	cmd.Flags().BoolVar(&flags.watermark, "watermark", false, "Include watermark")
	cmd.Flags().StringVarP(&flags.promptFile, "prompt-file", "f", "", "Read prompt from file")

	return cmd
}

func runImage2Video(cmd *cobra.Command, args []string, flags *image2videoFlags) error {
	// Validate image is required
	if flags.firstFrame == "" {
		return common.WriteError(cmd, "missing_image", "first frame image is required (-i)")
	}

	// Validate image file exists (if local)
	if !isURL(flags.firstFrame) {
		if _, err := os.Stat(flags.firstFrame); os.IsNotExist(err) {
			return common.WriteError(cmd, "image_not_found", fmt.Sprintf("image not found: %s", flags.firstFrame))
		}
	}

	// Validate last frame if provided
	if flags.lastFrame != "" && !isURL(flags.lastFrame) {
		if _, err := os.Stat(flags.lastFrame); os.IsNotExist(err) {
			return common.WriteError(cmd, "image_not_found", fmt.Sprintf("tail image not found: %s", flags.lastFrame))
		}
	}

	// Validate static mask if provided
	if flags.staticMask != "" && !isURL(flags.staticMask) {
		if _, err := os.Stat(flags.staticMask); os.IsNotExist(err) {
			return common.WriteError(cmd, "mask_not_found", fmt.Sprintf("static mask not found: %s", flags.staticMask))
		}
	}

	// Validate dynamic mask files if provided
	if flags.dynamicMask != "" {
		var masks []map[string]any
		if err := json.Unmarshal([]byte(flags.dynamicMask), &masks); err == nil {
			for i, m := range masks {
				if maskPath, ok := m["mask"].(string); ok && maskPath != "" && !isURL(maskPath) {
					if _, err := os.Stat(maskPath); os.IsNotExist(err) {
						return common.WriteError(cmd, "mask_not_found", fmt.Sprintf("dynamic mask %d not found: %s", i+1, maskPath))
					}
				}
			}
		}
	}

	// Get prompt (optional for image2video)
	prompt, _ := getPrompt(args, flags.promptFile, cmd.InOrStdin())

	// Validate model
	if !validI2VModels[flags.model] {
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

	// Validate cfg_scale
	if flags.cfgScale < 0 || flags.cfgScale > 1 {
		return common.WriteError(cmd, "invalid_cfg_scale", "cfg-scale must be between 0 and 1")
	}

	// Validate last frame (first+last frame) compatibility
	// Most models only support pro mode for first+last frame
	if flags.lastFrame != "" {
		switch flags.model {
		case "kling-v1":
			// kling-v1: supports std 5s and pro 5s (no 10s)
			if flags.duration != "5" {
				return common.WriteError(cmd, "incompatible_last_frame", "first+last frame only supports 5s duration for kling-v1")
			}
		case "kling-v2-master", "kling-v2-1-master":
			// These models don't support first+last frame
			return common.WriteError(cmd, "incompatible_last_frame", fmt.Sprintf("first+last frame not supported by %s", flags.model))
		default:
			// Other models: only pro mode
			if flags.mode != "pro" {
				return common.WriteError(cmd, "incompatible_last_frame", "first+last frame only supported in pro mode")
			}
		}
	}

	// Validate motion brush (static/dynamic mask) compatibility
	if flags.staticMask != "" || flags.dynamicMask != "" {
		switch flags.model {
		case "kling-v1":
			// kling-v1: supports std 5s and pro 5s
			if flags.duration != "5" {
				return common.WriteError(cmd, "incompatible_motion_brush", "motion brush only supports 5s duration for kling-v1")
			}
		case "kling-v1-5":
			// kling-v1-5: only pro 5s
			if flags.mode != "pro" || flags.duration != "5" {
				return common.WriteError(cmd, "incompatible_motion_brush", "motion brush only supported in pro mode with 5s duration for kling-v1-5")
			}
		default:
			// Other models don't support motion brush
			return common.WriteError(cmd, "incompatible_motion_brush", fmt.Sprintf("motion brush not supported by %s", flags.model))
		}
	}

	// Validate camera control compatibility
	// kling-v1-5: only pro 5s, simple type only
	if flags.cameraControl != "" {
		if flags.model != "kling-v1-5" {
			return common.WriteError(cmd, "incompatible_camera_control", "camera control for image2video only supported by kling-v1-5")
		}
		if flags.mode != "pro" || flags.duration != "5" {
			return common.WriteError(cmd, "incompatible_camera_control", "camera control only supported in pro mode with 5s duration")
		}
	}

	// Validate sound/voice compatibility (only v2.6)
	if flags.sound && flags.model != "kling-v2-6" {
		return common.WriteError(cmd, "incompatible_sound", "sound generation only supported by kling-v2-6")
	}
	if len(flags.voice) > 0 && flags.model != "kling-v2-6" {
		return common.WriteError(cmd, "incompatible_voice", "voice control only supported by kling-v2-6")
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

	// Resolve image URL
	imageURL, err := resolveImageURL(flags.firstFrame)
	if err != nil {
		return common.WriteError(cmd, "image_read_error", fmt.Sprintf("cannot read image: %s", err.Error()))
	}

	// Build request body
	body := map[string]any{
		"model_name": flags.model,
		"image":      imageURL,
		"mode":       flags.mode,
		"duration":   flags.duration,
	}

	if prompt != "" {
		body["prompt"] = prompt
	}

	if flags.negativePrompt != "" {
		body["negative_prompt"] = flags.negativePrompt
	}

	// Resolve last frame if provided
	if flags.lastFrame != "" {
		tailURL, err := resolveImageURL(flags.lastFrame)
		if err != nil {
			return common.WriteError(cmd, "image_read_error", fmt.Sprintf("cannot read tail image: %s", err.Error()))
		}
		body["image_tail"] = tailURL
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

	// Add static mask if provided
	if flags.staticMask != "" {
		maskURL, err := resolveImageURL(flags.staticMask)
		if err != nil {
			return common.WriteError(cmd, "mask_read_error", fmt.Sprintf("cannot read static mask: %s", err.Error()))
		}
		body["static_mask"] = maskURL
	}

	// Parse and add dynamic mask if provided
	if flags.dynamicMask != "" {
		var dynamicMasks []map[string]any
		if err := json.Unmarshal([]byte(flags.dynamicMask), &dynamicMasks); err != nil {
			return common.WriteError(cmd, "invalid_dynamic_mask", fmt.Sprintf("invalid dynamic mask JSON: %s", err.Error()))
		}
		// Process each mask - convert local file to base64
		for i, dm := range dynamicMasks {
			if maskPath, ok := dm["mask"].(string); ok && maskPath != "" {
				maskURL, err := resolveImageURL(maskPath)
				if err != nil {
					return common.WriteError(cmd, "mask_read_error", fmt.Sprintf("cannot read dynamic mask %d: %s", i+1, err.Error()))
				}
				dynamicMasks[i]["mask"] = maskURL
			}
		}
		body["dynamic_masks"] = dynamicMasks
	}

	// Add voice list if provided
	if len(flags.voice) > 0 {
		body["voice_list"] = flags.voice
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
	req, err := http.NewRequest("POST", getKlingAPIBase()+"/v1/videos/image2video", bytes.NewReader(jsonBody))
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
