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
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/cobra"
)

const (
	klingAPIBaseDefault = "https://api-beijing.klingai.com"
	klingModelO1        = "kling-video-o1"
)

// getKlingAPIBase returns the Kling API base URL.
// Priority: environment variable > config file > default
func getKlingAPIBase() string {
	if url := config.GetAPIKey("KLING_BASE_URL"); url != "" {
		return url
	}
	return klingAPIBaseDefault
}

var (
	validModes  = map[string]bool{"std": true, "pro": true}
	validRatios = map[string]bool{"16:9": true, "9:16": true, "1:1": true}
)

type createFlags struct {
	firstFrame      string
	lastFrame       string
	refImages       []string
	elements        []int64
	refVideo        string
	baseVideo       string
	refExcludeSound bool
	promptFile      string
	mode            string
	duration        int
	ratio           string
	watermark       bool
}

func newCreateCmd() *cobra.Command {
	flags := &createFlags{}

	cmd := &cobra.Command{
		Use:           "create [prompt]",
		Short:         "Create a video generation task",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.firstFrame, "first-frame", "i", "", "First frame image path")
	cmd.Flags().StringVar(&flags.lastFrame, "last-frame", "", "Last frame image path (requires --first-frame)")
	cmd.Flags().StringArrayVar(&flags.refImages, "ref-image", nil, "Reference image(s), use <<<image_N>>> in prompt")
	cmd.Flags().Int64SliceVar(&flags.elements, "element", nil, "Element ID(s), use <<<element_N>>> in prompt")
	cmd.Flags().StringVar(&flags.refVideo, "ref-video", "", "Reference video URL (style/camera reference)")
	cmd.Flags().StringVar(&flags.baseVideo, "base-video", "", "Base video URL for editing")
	cmd.Flags().BoolVar(&flags.refExcludeSound, "ref-exclude-sound", false, "Exclude sound from ref/base video")
	cmd.Flags().StringVarP(&flags.promptFile, "prompt-file", "f", "", "Read prompt from file")
	cmd.Flags().StringVar(&flags.mode, "mode", "pro", "Generation mode: std, pro")
	cmd.Flags().IntVarP(&flags.duration, "duration", "d", 5, "Video duration in seconds (3-10)")
	cmd.Flags().StringVarP(&flags.ratio, "ratio", "r", "16:9", "Aspect ratio: 16:9, 9:16, 1:1")
	cmd.Flags().BoolVar(&flags.watermark, "watermark", false, "Include watermark")

	return cmd
}

func runCreate(cmd *cobra.Command, args []string, flags *createFlags) error {
	// Get prompt
	prompt, err := getPrompt(args, flags.promptFile, cmd.InOrStdin())
	if err != nil {
		return common.WriteError(cmd, "missing_prompt", err.Error())
	}

	// Validate mode
	if !validModes[flags.mode] {
		return common.WriteError(cmd, "invalid_mode", fmt.Sprintf("invalid mode '%s', use std or pro", flags.mode))
	}

	// Validate ratio
	if !validRatios[flags.ratio] {
		return common.WriteError(cmd, "invalid_ratio", fmt.Sprintf("invalid ratio '%s', use 16:9, 9:16, or 1:1", flags.ratio))
	}

	// Validate duration
	if flags.duration < 3 || flags.duration > 10 {
		return common.WriteError(cmd, "invalid_duration", "duration must be between 3 and 10 seconds")
	}

	// Validate first frame (only check local files, URLs are validated by API)
	if flags.firstFrame != "" && !isURL(flags.firstFrame) {
		if _, err := os.Stat(flags.firstFrame); os.IsNotExist(err) {
			return common.WriteError(cmd, "frame_not_found", fmt.Sprintf("first frame image not found: %s", flags.firstFrame))
		}
	}

	// Validate last frame
	if flags.lastFrame != "" {
		if flags.firstFrame == "" {
			return common.WriteError(cmd, "last_frame_requires_first", "--last-frame requires --first-frame")
		}
		if !isURL(flags.lastFrame) {
			if _, err := os.Stat(flags.lastFrame); os.IsNotExist(err) {
				return common.WriteError(cmd, "frame_not_found", fmt.Sprintf("last frame image not found: %s", flags.lastFrame))
			}
		}
	}

	// Validate ref images (only check local files)
	for _, img := range flags.refImages {
		if !isURL(img) {
			if _, err := os.Stat(img); os.IsNotExist(err) {
				return common.WriteError(cmd, "ref_image_not_found", fmt.Sprintf("reference image not found: %s", img))
			}
		}
	}

	// Validate video flags are not both set
	if flags.refVideo != "" && flags.baseVideo != "" {
		return common.WriteError(cmd, "conflicting_video_flags", "cannot use --ref-video and --base-video together")
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
		"model_name": klingModelO1,
		"prompt":     prompt,
		"mode":       flags.mode,
		"duration":   fmt.Sprintf("%d", flags.duration),
	}

	// Add aspect ratio (only for text-to-video or ref images without first frame)
	if flags.firstFrame == "" && flags.baseVideo == "" {
		body["aspect_ratio"] = flags.ratio
	}

	// Build image_list
	imageList := []map[string]any{}

	// Add first frame
	if flags.firstFrame != "" {
		imgURL, err := resolveImageURL(flags.firstFrame)
		if err != nil {
			return common.WriteError(cmd, "frame_read_error", fmt.Sprintf("cannot read first frame: %s", err.Error()))
		}
		imageList = append(imageList, map[string]any{
			"image_url": imgURL,
			"type":      "first_frame",
		})
	}

	// Add last frame
	if flags.lastFrame != "" {
		imgURL, err := resolveImageURL(flags.lastFrame)
		if err != nil {
			return common.WriteError(cmd, "frame_read_error", fmt.Sprintf("cannot read last frame: %s", err.Error()))
		}
		imageList = append(imageList, map[string]any{
			"image_url": imgURL,
			"type":      "end_frame",
		})
	}

	// Add reference images
	for _, img := range flags.refImages {
		imgURL, err := resolveImageURL(img)
		if err != nil {
			return common.WriteError(cmd, "ref_image_read_error", fmt.Sprintf("cannot read reference image: %s", err.Error()))
		}
		imageList = append(imageList, map[string]any{
			"image_url": imgURL,
		})
	}

	if len(imageList) > 0 {
		body["image_list"] = imageList
	}

	// Build element_list
	if len(flags.elements) > 0 {
		elementList := []map[string]any{}
		for _, id := range flags.elements {
			elementList = append(elementList, map[string]any{
				"element_id": id,
			})
		}
		body["element_list"] = elementList
	}

	// Build video_list
	if flags.refVideo != "" || flags.baseVideo != "" {
		videoURL := flags.refVideo
		referType := "feature"
		if flags.baseVideo != "" {
			videoURL = flags.baseVideo
			referType = "base"
		}

		keepSound := "yes"
		if flags.refExcludeSound {
			keepSound = "no"
		}

		body["video_list"] = []map[string]any{
			{
				"video_url":           videoURL,
				"refer_type":          referType,
				"keep_original_sound": keepSound,
			},
		}
	}

	// Add watermark
	if flags.watermark {
		body["watermark_info"] = map[string]bool{"enabled": true}
	}

	// Serialize request
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot serialize request: %s", err.Error()))
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", getKlingAPIBase()+"/v1/videos/omni-video", bytes.NewReader(jsonBody))
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

func getPrompt(args []string, filePath string, stdin io.Reader) (string, error) {
	// Priority 1: Positional argument
	if len(args) > 0 {
		text := strings.TrimSpace(args[0])
		if text != "" {
			return text, nil
		}
	}

	// Priority 2: File
	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("cannot read file: %s", err.Error())
		}
		text := strings.TrimSpace(string(data))
		if text == "" {
			return "", fmt.Errorf("file is empty")
		}
		return text, nil
	}

	// Priority 3: Stdin (only if not a terminal)
	if stdin != nil {
		if f, ok := stdin.(*os.File); ok {
			stat, _ := f.Stat()
			if (stat.Mode() & os.ModeCharDevice) != 0 {
				return "", fmt.Errorf("no prompt provided")
			}
		}
		data, err := io.ReadAll(stdin)
		if err != nil {
			return "", fmt.Errorf("cannot read stdin: %s", err.Error())
		}
		text := strings.TrimSpace(string(data))
		if text == "" {
			return "", fmt.Errorf("stdin is empty")
		}
		return text, nil
	}

	return "", fmt.Errorf("no prompt provided")
}

func generateJWT(accessKey, secretKey string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"iss": accessKey,
		"exp": now.Add(30 * time.Minute).Unix(),
		"nbf": now.Add(-5 * time.Second).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

// isURL checks if the given string is a URL
func isURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

// resolveImageURL returns the image URL for API request.
// If input is a URL, returns it directly.
// If input is a local file path, encodes it as pure base64 (Kling API requires no data URL prefix).
func resolveImageURL(input string) (string, error) {
	if isURL(input) {
		return input, nil
	}
	return encodeImageToBase64(input)
}

// encodeImageToBase64 encodes local image file to pure base64 string.
// Kling API requires base64 without "data:image/...;base64," prefix.
func encodeImageToBase64(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func handleAPIError(cmd *cobra.Command, err error) error {
	errStr := err.Error()

	if strings.Contains(errStr, "timeout") {
		return common.WriteError(cmd, "timeout", "request timed out")
	}
	if strings.Contains(errStr, "connection") || strings.Contains(errStr, "refused") {
		return common.WriteError(cmd, "connection_error", "cannot connect to Kling API")
	}
	if strings.Contains(errStr, "no such host") || strings.Contains(errStr, "dns") {
		return common.WriteError(cmd, "connection_error", "cannot resolve Kling API host")
	}

	return common.WriteError(cmd, "api_error", err.Error())
}

func handleKlingError(cmd *cobra.Command, code int, message string) error {
	switch code {
	case 1000, 1001, 1002, 1003, 1004:
		return common.WriteError(cmd, "invalid_api_key", message)
	case 1100, 1101, 1102:
		return common.WriteError(cmd, "quota_exceeded", message)
	case 1103:
		return common.WriteError(cmd, "permission_denied", message)
	case 1200, 1201:
		return common.WriteError(cmd, "invalid_request", message)
	case 1202, 1203:
		return common.WriteError(cmd, "task_not_found", message)
	case 1300, 1301:
		return common.WriteError(cmd, "content_policy", message)
	case 1302, 1303:
		return common.WriteError(cmd, "rate_limit", message)
	case 5000, 5001, 5002:
		return common.WriteError(cmd, "server_error", message)
	default:
		return common.WriteError(cmd, "api_error", message)
	}
}
