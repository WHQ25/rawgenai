package image

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/cli/kling/video"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

var validModels = map[string]bool{
	"kling-v1":     true,
	"kling-v1-5":   true,
	"kling-v2":     true,
	"kling-v2-new": true,
	"kling-v2-1":   true,
}

var validRatios = map[string]bool{
	"16:9": true,
	"9:16": true,
	"1:1":  true,
	"4:3":  true,
	"3:4":  true,
	"3:2":  true,
	"2:3":  true,
	"21:9": true,
}

var validResolutions = map[string]bool{
	"1k": true,
	"2k": true,
}

var validImageReferences = map[string]bool{
	"subject": true,
	"face":    true,
}

type createFlags struct {
	image          string
	imageReference string
	imageFidelity  float64
	humanFidelity  float64
	negativePrompt string
	model          string
	resolution     string
	count          int
	aspectRatio    string
	watermark      bool
	promptFile     string
	callbackURL    string
	externalTaskID string
}

func newCreateCmd() *cobra.Command {
	flags := &createFlags{}

	cmd := &cobra.Command{
		Use:           "create [prompt]",
		Short:         "Create an image generation task",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(cmd, args, flags)
		},
	}

	cmd.Flags().StringVar(&flags.image, "image", "", "Reference image (local file or URL)")
	cmd.Flags().StringVar(&flags.imageReference, "image-reference", "", "Image reference type: subject, face (kling-v1-5 only)")
	cmd.Flags().Float64Var(&flags.imageFidelity, "image-fidelity", 0.5, "Image reference strength (0-1)")
	cmd.Flags().Float64Var(&flags.humanFidelity, "human-fidelity", 0.45, "Face reference strength (0-1), only for image-reference=subject")
	cmd.Flags().StringVar(&flags.negativePrompt, "negative", "", "Negative prompt (not supported for image input)")
	cmd.Flags().StringVarP(&flags.model, "model", "m", "kling-v1", "Model: kling-v1, kling-v1-5, kling-v2, kling-v2-new, kling-v2-1")
	cmd.Flags().StringVar(&flags.resolution, "resolution", "1k", "Resolution: 1k, 2k")
	cmd.Flags().IntVarP(&flags.count, "count", "n", 1, "Number of images (1-9)")
	cmd.Flags().StringVarP(&flags.aspectRatio, "ratio", "r", "16:9", "Aspect ratio: 16:9, 9:16, 1:1, 4:3, 3:4, 3:2, 2:3, 21:9")
	cmd.Flags().BoolVar(&flags.watermark, "watermark", false, "Include watermark")
	cmd.Flags().StringVarP(&flags.promptFile, "prompt-file", "f", "", "Read prompt from file")
	cmd.Flags().StringVar(&flags.callbackURL, "callback-url", "", "Callback URL for task status changes")
	cmd.Flags().StringVar(&flags.externalTaskID, "external-task-id", "", "External task ID (must be unique)")

	return cmd
}

func runCreate(cmd *cobra.Command, args []string, flags *createFlags) error {
	// Get prompt
	prompt, err := video.GetPrompt(args, flags.promptFile, cmd.InOrStdin())
	if err != nil {
		return common.WriteError(cmd, "missing_prompt", err.Error())
	}

	// Validate model
	if !validModels[flags.model] {
		return common.WriteError(cmd, "invalid_model", "invalid model")
	}

	// Validate resolution
	if !validResolutions[flags.resolution] {
		return common.WriteError(cmd, "invalid_resolution", "resolution must be 1k or 2k")
	}

	// Validate aspect ratio
	if !validRatios[flags.aspectRatio] {
		return common.WriteError(cmd, "invalid_ratio", "invalid aspect ratio")
	}

	// Validate count
	if flags.count < 1 || flags.count > 9 {
		return common.WriteError(cmd, "invalid_count", "n must be between 1 and 9")
	}

	// Validate image reference
	if strings.TrimSpace(flags.imageReference) != "" && !validImageReferences[flags.imageReference] {
		return common.WriteError(cmd, "invalid_image_reference", "image-reference must be subject or face")
	}

	// Negative prompt not supported for image input
	if flags.image != "" && strings.TrimSpace(flags.negativePrompt) != "" {
		return common.WriteError(cmd, "negative_not_supported", "negative prompt is not supported when --image is provided")
	}

	// image-reference requires image
	if flags.imageReference != "" && flags.image == "" {
		return common.WriteError(cmd, "image_reference_requires_image", "--image-reference requires --image")
	}

	// image-reference only supported by kling-v1-5
	if flags.imageReference != "" && flags.model != "kling-v1-5" {
		return common.WriteError(cmd, "image_reference_model", "image-reference is only supported by kling-v1-5")
	}

	imageFidelitySet := cmd.Flags().Changed("image-fidelity")
	humanFidelitySet := cmd.Flags().Changed("human-fidelity")

	if (imageFidelitySet || humanFidelitySet) && flags.image == "" {
		return common.WriteError(cmd, "image_fidelity_requires_image", "image-fidelity/human-fidelity requires --image")
	}

	if imageFidelitySet && (flags.imageFidelity < 0 || flags.imageFidelity > 1) {
		return common.WriteError(cmd, "invalid_image_fidelity", "image-fidelity must be between 0 and 1")
	}

	if humanFidelitySet && (flags.humanFidelity < 0 || flags.humanFidelity > 1) {
		return common.WriteError(cmd, "invalid_human_fidelity", "human-fidelity must be between 0 and 1")
	}

	if humanFidelitySet && flags.imageReference != "subject" {
		return common.WriteError(cmd, "human_fidelity_requires_subject", "human-fidelity only works with image-reference=subject")
	}

	// Check API keys
	accessKey := config.GetAPIKey("KLING_ACCESS_KEY")
	secretKey := config.GetAPIKey("KLING_SECRET_KEY")
	if accessKey == "" || secretKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("KLING_ACCESS_KEY")+" and "+config.GetMissingKeyMessage("KLING_SECRET_KEY"))
	}

	// Generate JWT token
	token, err := video.GenerateJWT(accessKey, secretKey)
	if err != nil {
		return common.WriteError(cmd, "auth_error", fmt.Sprintf("failed to generate JWT: %s", err.Error()))
	}

	// Build request body
	body := map[string]any{
		"model_name":   flags.model,
		"prompt":       prompt,
		"aspect_ratio": flags.aspectRatio,
		"resolution":   flags.resolution,
		"n":            flags.count,
	}

	if flags.negativePrompt != "" {
		body["negative_prompt"] = flags.negativePrompt
	}

	if flags.image != "" {
		imageURL, err := video.ResolveImageURL(flags.image)
		if err != nil {
			if os.IsNotExist(err) {
				return common.WriteError(cmd, "image_not_found", "image file not found: "+flags.image)
			}
			return common.WriteError(cmd, "image_read_error", fmt.Sprintf("cannot read image: %s", err.Error()))
		}
		body["image"] = imageURL
	}

	if flags.imageReference != "" {
		body["image_reference"] = flags.imageReference
	}

	if imageFidelitySet {
		body["image_fidelity"] = flags.imageFidelity
	}
	if humanFidelitySet {
		body["human_fidelity"] = flags.humanFidelity
	}

	if flags.watermark {
		body["watermark_info"] = map[string]bool{"enabled": true}
	}

	if flags.callbackURL != "" {
		body["callback_url"] = flags.callbackURL
	}

	if flags.externalTaskID != "" {
		body["external_task_id"] = flags.externalTaskID
	}

	// Serialize request
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot serialize request: %s", err.Error()))
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", video.GetKlingAPIBase()+"/v1/images/generations", bytes.NewReader(jsonBody))
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot create request: %s", err.Error()))
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Send request
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return video.HandleAPIError(cmd, err)
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
			TaskInfo   *struct {
				ExternalTaskID string `json:"external_task_id"`
			} `json:"task_info"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot parse response: %s", err.Error()))
	}

	if result.Code != 0 {
		return video.HandleKlingError(cmd, result.Code, result.Message)
	}

	if result.Data == nil {
		return common.WriteError(cmd, "response_error", "no data in response")
	}

	output := map[string]any{
		"success": true,
		"task_id": result.Data.TaskID,
		"status":  result.Data.TaskStatus,
	}

	if result.Data.TaskInfo != nil && result.Data.TaskInfo.ExternalTaskID != "" {
		output["external_task_id"] = result.Data.TaskInfo.ExternalTaskID
	}

	return common.WriteSuccess(cmd, output)
}
