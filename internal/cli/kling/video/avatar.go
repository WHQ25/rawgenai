package video

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

type avatarFlags struct {
	image      string
	audio      string
	audioID    string
	mode       string
	watermark  bool
	promptFile string
}

func newAvatarCmd() *cobra.Command {
	flags := &avatarFlags{}

	cmd := &cobra.Command{
		Use:           "create-avatar [prompt]",
		Short:         "Create digital avatar video with lip sync",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAvatar(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.image, "image", "i", "", "Avatar reference image (required)")
	cmd.Flags().StringVarP(&flags.audio, "audio", "a", "", "Audio file for lip sync (local file or URL)")
	cmd.Flags().StringVar(&flags.audioID, "audio-id", "", "Audio ID from TTS preview (alternative to --audio)")
	cmd.Flags().StringVarP(&flags.mode, "mode", "m", "std", "Generation mode: std, pro")
	cmd.Flags().BoolVar(&flags.watermark, "watermark", false, "Include watermark")
	cmd.Flags().StringVarP(&flags.promptFile, "prompt-file", "f", "", "Read prompt from file")

	return cmd
}

func runAvatar(cmd *cobra.Command, args []string, flags *avatarFlags) error {
	// Validate image is required
	if flags.image == "" {
		return common.WriteError(cmd, "missing_image", "avatar reference image is required (-i)")
	}

	// Validate audio source (audio or audio-id required, but not both)
	if flags.audio == "" && flags.audioID == "" {
		return common.WriteError(cmd, "missing_audio", "audio file (--audio) or audio ID (--audio-id) is required")
	}
	if flags.audio != "" && flags.audioID != "" {
		return common.WriteError(cmd, "conflicting_audio", "cannot use both --audio and --audio-id")
	}

	// Validate image file exists (if local)
	if !isURL(flags.image) {
		if _, err := os.Stat(flags.image); os.IsNotExist(err) {
			return common.WriteError(cmd, "image_not_found", fmt.Sprintf("image not found: %s", flags.image))
		}
	}

	// Validate audio file exists (if local)
	if flags.audio != "" && !isURL(flags.audio) {
		if _, err := os.Stat(flags.audio); os.IsNotExist(err) {
			return common.WriteError(cmd, "audio_not_found", fmt.Sprintf("audio not found: %s", flags.audio))
		}
	}

	// Get prompt (optional)
	prompt, _ := getPrompt(args, flags.promptFile, cmd.InOrStdin())

	// Validate mode
	if !validModes[flags.mode] {
		return common.WriteError(cmd, "invalid_mode", fmt.Sprintf("invalid mode '%s', use std or pro", flags.mode))
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
	imageURL, err := resolveImageURL(flags.image)
	if err != nil {
		return common.WriteError(cmd, "image_read_error", fmt.Sprintf("cannot read image: %s", err.Error()))
	}

	// Build request body
	body := map[string]any{
		"image": imageURL,
		"mode":  flags.mode,
	}

	// Add audio source
	if flags.audioID != "" {
		body["audio_id"] = flags.audioID
	} else {
		// Resolve audio file
		audioURL, err := resolveAudioURL(flags.audio)
		if err != nil {
			return common.WriteError(cmd, "audio_read_error", fmt.Sprintf("cannot read audio: %s", err.Error()))
		}
		body["sound_file"] = audioURL
	}

	if prompt != "" {
		body["prompt"] = prompt
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
	req, err := http.NewRequest("POST", getKlingAPIBase()+"/v1/videos/avatar/image2video", bytes.NewReader(jsonBody))
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

// resolveAudioURL returns the audio URL for API request.
// If input is URL, returns as-is. If local file, encodes to base64.
func resolveAudioURL(input string) (string, error) {
	if isURL(input) {
		return input, nil
	}
	return encodeAudioToBase64(input)
}

// encodeAudioToBase64 encodes local audio file to pure base64 string.
func encodeAudioToBase64(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

// isAudioFile checks if the file has a supported audio extension.
func isAudioFile(path string) bool {
	lower := strings.ToLower(path)
	return strings.HasSuffix(lower, ".mp3") ||
		strings.HasSuffix(lower, ".wav") ||
		strings.HasSuffix(lower, ".m4a") ||
		strings.HasSuffix(lower, ".aac")
}
