package seed

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

const (
	seedVideoModelID = "doubao-seedance-1-5-pro-251215"
	arkAPIBase       = "https://ark.cn-beijing.volces.com/api/v3"
)

// Valid values
var (
	validRatios = map[string]bool{
		"16:9": true,
		"9:16": true,
		"4:3":  true,
		"3:4":  true,
		"1:1":  true,
		"21:9": true,
	}
	validResolutions = map[string]bool{
		"480p":  true,
		"720p":  true,
		"1080p": true,
	}
	validStatuses = map[string]bool{
		"queued":    true,
		"running":   true,
		"succeeded": true,
		"failed":    true,
	}
)

// Flag structs
type videoCreateFlags struct {
	promptFile      string
	firstFrame      string
	lastFrame       string
	ratio           string
	resolution      string
	duration        int
	audio           bool
	seed            int
	watermark       bool
	returnLastFrame bool
}

type videoDownloadFlags struct {
	output    string
	lastFrame string
}

type videoListFlags struct {
	limit  int
	status string
}

// Commands
var videoCmd = newVideoCmd()

func newVideoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "video",
		Short: "Generate videos using Seedance 1.5 Pro",
		Long:  "Generate videos using ByteDance Seedance 1.5 Pro model.",
	}

	cmd.AddCommand(newVideoCreateCmd())
	cmd.AddCommand(newVideoStatusCmd())
	cmd.AddCommand(newVideoDownloadCmd())
	cmd.AddCommand(newVideoListCmd())
	cmd.AddCommand(newVideoDeleteCmd())

	return cmd
}

// ===== Create Command =====

func newVideoCreateCmd() *cobra.Command {
	flags := &videoCreateFlags{}

	cmd := &cobra.Command{
		Use:           "create [prompt]",
		Short:         "Create a video generation task",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVideoCreate(cmd, args, flags)
		},
	}

	cmd.Flags().StringVar(&flags.promptFile, "prompt-file", "", "Input prompt file")
	cmd.Flags().StringVar(&flags.firstFrame, "first-frame", "", "First frame image (JPEG/PNG/WebP)")
	cmd.Flags().StringVar(&flags.lastFrame, "last-frame", "", "Last frame image (requires --first-frame)")
	cmd.Flags().StringVarP(&flags.ratio, "ratio", "r", "16:9", "Aspect ratio: 16:9, 9:16, 4:3, 3:4, 1:1, 21:9")
	cmd.Flags().StringVar(&flags.resolution, "resolution", "1080p", "Resolution: 480p, 720p, 1080p")
	cmd.Flags().IntVarP(&flags.duration, "duration", "d", 5, "Duration in seconds (4-12)")
	cmd.Flags().BoolVar(&flags.audio, "audio", false, "Generate video with audio")
	cmd.Flags().IntVar(&flags.seed, "seed", 0, "Random seed for reproducibility")
	cmd.Flags().BoolVar(&flags.watermark, "watermark", false, "Add watermark to output")
	cmd.Flags().BoolVar(&flags.returnLastFrame, "return-last-frame", false, "Return last frame URL (for chaining)")

	return cmd
}

func runVideoCreate(cmd *cobra.Command, args []string, flags *videoCreateFlags) error {
	// Get prompt
	prompt, err := getVideoPrompt(args, flags.promptFile, cmd.InOrStdin())
	if err != nil {
		return common.WriteError(cmd, "missing_prompt", err.Error())
	}

	// Validate ratio
	if !validRatios[flags.ratio] {
		return common.WriteError(cmd, "invalid_ratio", fmt.Sprintf("invalid ratio '%s', use 16:9, 9:16, 4:3, 3:4, 1:1, or 21:9", flags.ratio))
	}

	// Validate resolution
	if !validResolutions[flags.resolution] {
		return common.WriteError(cmd, "invalid_resolution", fmt.Sprintf("invalid resolution '%s', use 480p, 720p, or 1080p", flags.resolution))
	}

	// Validate duration
	if flags.duration < 4 || flags.duration > 12 {
		return common.WriteError(cmd, "invalid_duration", "duration must be between 4 and 12 seconds")
	}

	// Validate first frame
	if flags.firstFrame != "" {
		if _, err := os.Stat(flags.firstFrame); os.IsNotExist(err) {
			return common.WriteError(cmd, "first_frame_not_found", fmt.Sprintf("first frame image not found: %s", flags.firstFrame))
		}
	}

	// Validate last frame
	if flags.lastFrame != "" {
		if flags.firstFrame == "" {
			return common.WriteError(cmd, "last_frame_requires_first", "--last-frame requires --first-frame")
		}
		if _, err := os.Stat(flags.lastFrame); os.IsNotExist(err) {
			return common.WriteError(cmd, "last_frame_not_found", fmt.Sprintf("last frame image not found: %s", flags.lastFrame))
		}
	}

	// Check API key
	apiKey := config.GetAPIKey("ARK_API_KEY")
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("ARK_API_KEY"))
	}

	// Build content array
	content := []map[string]any{
		{"type": "text", "text": prompt},
	}

	// Add first frame if provided
	if flags.firstFrame != "" {
		imgData, err := os.ReadFile(flags.firstFrame)
		if err != nil {
			return common.WriteError(cmd, "image_read_error", fmt.Sprintf("cannot read first frame: %s", err.Error()))
		}
		mimeType := getSeedMimeType(flags.firstFrame)
		dataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(imgData))
		content = append(content, map[string]any{
			"type":      "image_url",
			"image_url": map[string]string{"url": dataURL},
			"role":      "first_frame",
		})
	}

	// Add last frame if provided
	if flags.lastFrame != "" {
		imgData, err := os.ReadFile(flags.lastFrame)
		if err != nil {
			return common.WriteError(cmd, "image_read_error", fmt.Sprintf("cannot read last frame: %s", err.Error()))
		}
		mimeType := getSeedMimeType(flags.lastFrame)
		dataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(imgData))
		content = append(content, map[string]any{
			"type":      "image_url",
			"image_url": map[string]string{"url": dataURL},
			"role":      "last_frame",
		})
	}

	// Build request body
	body := map[string]any{
		"model":            seedVideoModelID,
		"content":          content,
		"ratio":            flags.ratio,
		"resolution":       flags.resolution,
		"duration":         flags.duration,
		"generate_audio":   flags.audio,
		"watermark":        flags.watermark,
		"return_last_frame": flags.returnLastFrame,
	}

	if flags.seed != 0 {
		body["seed"] = flags.seed
	}

	// Serialize request
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot serialize request: %s", err.Error()))
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", arkAPIBase+"/contents/generations/tasks", bytes.NewReader(jsonBody))
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot create request: %s", err.Error()))
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return handleVideoAPIError(cmd, err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot read response: %s", err.Error()))
	}

	// Parse response
	var result struct {
		ID    string `json:"id"`
		Error *struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot parse response: %s", err.Error()))
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		if result.Error != nil {
			return common.WriteError(cmd, result.Error.Code, result.Error.Message)
		}
		return handleVideoHTTPError(cmd, resp.StatusCode, string(respBody))
	}

	if result.Error != nil {
		return common.WriteError(cmd, result.Error.Code, result.Error.Message)
	}

	// Return success
	return common.WriteSuccess(cmd, map[string]any{
		"success": true,
		"task_id": result.ID,
		"status":  "queued",
	})
}

// ===== Status Command =====

func newVideoStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "status <task_id>",
		Short:         "Get video generation status",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVideoStatus(cmd, args)
		},
	}

	return cmd
}

func runVideoStatus(cmd *cobra.Command, args []string) error {
	// Validate task ID
	if len(args) == 0 || strings.TrimSpace(args[0]) == "" {
		return common.WriteError(cmd, "missing_task_id", "task ID is required")
	}
	taskID := args[0]

	// Check API key
	apiKey := config.GetAPIKey("ARK_API_KEY")
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("ARK_API_KEY"))
	}

	// Create HTTP request
	req, err := http.NewRequest("GET", arkAPIBase+"/contents/generations/tasks/"+taskID, nil)
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot create request: %s", err.Error()))
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return handleVideoAPIError(cmd, err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot read response: %s", err.Error()))
	}

	// Parse response
	var result struct {
		ID         string `json:"id"`
		Status     string `json:"status"`
		Content    *struct {
			VideoURL     string `json:"video_url"`
			LastFrameURL string `json:"last_frame_url"`
		} `json:"content"`
		Resolution string `json:"resolution"`
		Ratio      string `json:"ratio"`
		Duration   int    `json:"duration"`
		Seed       int    `json:"seed"`
		Error      *struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot parse response: %s", err.Error()))
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		if result.Error != nil {
			return common.WriteError(cmd, result.Error.Code, result.Error.Message)
		}
		return handleVideoHTTPError(cmd, resp.StatusCode, string(respBody))
	}

	// Handle failed status
	if result.Status == "failed" {
		if result.Error != nil {
			return common.WriteError(cmd, result.Error.Code, result.Error.Message)
		}
		return common.WriteError(cmd, "video_failed", "video generation failed")
	}

	// Build response
	output := map[string]any{
		"success": true,
		"task_id": result.ID,
		"status":  result.Status,
	}

	if result.Status == "succeeded" && result.Content != nil {
		output["video_url"] = result.Content.VideoURL
		if result.Content.LastFrameURL != "" {
			output["last_frame_url"] = result.Content.LastFrameURL
		}
		output["resolution"] = result.Resolution
		output["ratio"] = result.Ratio
		output["duration"] = result.Duration
		if result.Seed != 0 {
			output["seed"] = result.Seed
		}
	}

	return common.WriteSuccess(cmd, output)
}

// ===== Download Command =====

func newVideoDownloadCmd() *cobra.Command {
	flags := &videoDownloadFlags{}

	cmd := &cobra.Command{
		Use:           "download <task_id>",
		Short:         "Download completed video",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVideoDownload(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.output, "output", "o", "", "Output file path (.mp4)")
	cmd.Flags().StringVar(&flags.lastFrame, "last-frame", "", "Also save last frame to this path")

	return cmd
}

func runVideoDownload(cmd *cobra.Command, args []string, flags *videoDownloadFlags) error {
	// Validate task ID
	if len(args) == 0 || strings.TrimSpace(args[0]) == "" {
		return common.WriteError(cmd, "missing_task_id", "task ID is required")
	}
	taskID := args[0]

	// Validate output
	if flags.output == "" {
		return common.WriteError(cmd, "missing_output", "output file is required, use -o flag")
	}

	// Validate format
	ext := strings.ToLower(filepath.Ext(flags.output))
	if ext != ".mp4" {
		return common.WriteError(cmd, "invalid_format", fmt.Sprintf("unsupported format '%s', only .mp4 is supported", ext))
	}

	// Check API key
	apiKey := config.GetAPIKey("ARK_API_KEY")
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("ARK_API_KEY"))
	}

	// Get task status first
	req, err := http.NewRequest("GET", arkAPIBase+"/contents/generations/tasks/"+taskID, nil)
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot create request: %s", err.Error()))
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return handleVideoAPIError(cmd, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot read response: %s", err.Error()))
	}

	var result struct {
		ID      string `json:"id"`
		Status  string `json:"status"`
		Content *struct {
			VideoURL     string `json:"video_url"`
			LastFrameURL string `json:"last_frame_url"`
		} `json:"content"`
		Error *struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot parse response: %s", err.Error()))
	}

	if resp.StatusCode != http.StatusOK {
		if result.Error != nil {
			return common.WriteError(cmd, result.Error.Code, result.Error.Message)
		}
		return handleVideoHTTPError(cmd, resp.StatusCode, string(respBody))
	}

	// Check status
	if result.Status == "failed" {
		if result.Error != nil {
			return common.WriteError(cmd, result.Error.Code, result.Error.Message)
		}
		return common.WriteError(cmd, "video_failed", "video generation failed")
	}

	if result.Status != "succeeded" {
		return common.WriteError(cmd, "video_not_ready", fmt.Sprintf("video is not ready, current status: %s", result.Status))
	}

	if result.Content == nil || result.Content.VideoURL == "" {
		return common.WriteError(cmd, "no_video", "no video URL in response")
	}

	// Download video
	videoResp, err := http.Get(result.Content.VideoURL)
	if err != nil {
		return common.WriteError(cmd, "download_error", fmt.Sprintf("cannot download video: %s", err.Error()))
	}
	defer videoResp.Body.Close()

	if videoResp.StatusCode != http.StatusOK {
		return common.WriteError(cmd, "download_error", fmt.Sprintf("download failed with status: %d", videoResp.StatusCode))
	}

	videoData, err := io.ReadAll(videoResp.Body)
	if err != nil {
		return common.WriteError(cmd, "download_error", fmt.Sprintf("cannot read video data: %s", err.Error()))
	}

	// Save video
	absPath, err := filepath.Abs(flags.output)
	if err != nil {
		absPath = flags.output
	}

	if err := os.WriteFile(absPath, videoData, 0644); err != nil {
		return common.WriteError(cmd, "output_write_error", fmt.Sprintf("cannot write output file: %s", err.Error()))
	}

	// Build response
	output := map[string]any{
		"success": true,
		"task_id": taskID,
		"file":    absPath,
	}

	// Download last frame if requested
	if flags.lastFrame != "" && result.Content.LastFrameURL != "" {
		lastFrameResp, err := http.Get(result.Content.LastFrameURL)
		if err == nil {
			defer lastFrameResp.Body.Close()
			if lastFrameResp.StatusCode == http.StatusOK {
				lastFrameData, err := io.ReadAll(lastFrameResp.Body)
				if err == nil {
					lastFramePath, _ := filepath.Abs(flags.lastFrame)
					if err := os.WriteFile(lastFramePath, lastFrameData, 0644); err == nil {
						output["last_frame_file"] = lastFramePath
					}
				}
			}
		}
	}

	return common.WriteSuccess(cmd, output)
}

// ===== List Command =====

func newVideoListCmd() *cobra.Command {
	flags := &videoListFlags{}

	cmd := &cobra.Command{
		Use:           "list",
		Short:         "List video generation tasks",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVideoList(cmd, flags)
		},
	}

	cmd.Flags().IntVarP(&flags.limit, "limit", "l", 20, "Maximum number of tasks to return (1-100)")
	cmd.Flags().StringVarP(&flags.status, "status", "s", "", "Filter by status: queued, running, succeeded, failed")

	return cmd
}

func runVideoList(cmd *cobra.Command, flags *videoListFlags) error {
	// Validate limit
	if flags.limit < 1 || flags.limit > 100 {
		return common.WriteError(cmd, "invalid_limit", "limit must be between 1 and 100")
	}

	// Validate status
	if flags.status != "" && !validStatuses[flags.status] {
		return common.WriteError(cmd, "invalid_status", fmt.Sprintf("invalid status '%s', use queued, running, succeeded, or failed", flags.status))
	}

	// Check API key
	apiKey := config.GetAPIKey("ARK_API_KEY")
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("ARK_API_KEY"))
	}

	// Build URL with query params
	url := fmt.Sprintf("%s/contents/generations/tasks?limit=%d", arkAPIBase, flags.limit)
	if flags.status != "" {
		url += "&status=" + flags.status
	}

	// Create HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot create request: %s", err.Error()))
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return handleVideoAPIError(cmd, err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot read response: %s", err.Error()))
	}

	// Parse response
	var result struct {
		Data []struct {
			ID        string `json:"id"`
			Status    string `json:"status"`
			CreatedAt int64  `json:"created_at"`
		} `json:"data"`
		Error *struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot parse response: %s", err.Error()))
	}

	if resp.StatusCode != http.StatusOK {
		if result.Error != nil {
			return common.WriteError(cmd, result.Error.Code, result.Error.Message)
		}
		return handleVideoHTTPError(cmd, resp.StatusCode, string(respBody))
	}

	// Build response
	tasks := make([]map[string]any, 0, len(result.Data))
	for _, task := range result.Data {
		tasks = append(tasks, map[string]any{
			"task_id":    task.ID,
			"status":     task.Status,
			"created_at": task.CreatedAt,
		})
	}

	return common.WriteSuccess(cmd, map[string]any{
		"success": true,
		"tasks":   tasks,
		"count":   len(tasks),
	})
}

// ===== Delete Command =====

func newVideoDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "delete <task_id>",
		Short:         "Delete or cancel a video generation task",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVideoDelete(cmd, args)
		},
	}

	return cmd
}

func runVideoDelete(cmd *cobra.Command, args []string) error {
	// Validate task ID
	if len(args) == 0 || strings.TrimSpace(args[0]) == "" {
		return common.WriteError(cmd, "missing_task_id", "task ID is required")
	}
	taskID := args[0]

	// Check API key
	apiKey := config.GetAPIKey("ARK_API_KEY")
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("ARK_API_KEY"))
	}

	// Create HTTP request
	req, err := http.NewRequest("DELETE", arkAPIBase+"/contents/generations/tasks/"+taskID, nil)
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot create request: %s", err.Error()))
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return handleVideoAPIError(cmd, err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot read response: %s", err.Error()))
	}

	// Parse response for errors
	var result struct {
		Error *struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(respBody, &result); err == nil && result.Error != nil {
		return common.WriteError(cmd, result.Error.Code, result.Error.Message)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return handleVideoHTTPError(cmd, resp.StatusCode, string(respBody))
	}

	return common.WriteSuccess(cmd, map[string]any{
		"success": true,
		"task_id": taskID,
		"deleted": true,
	})
}

// ===== Helper Functions =====

func getVideoPrompt(args []string, filePath string, stdin io.Reader) (string, error) {
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

func handleVideoAPIError(cmd *cobra.Command, err error) error {
	errStr := err.Error()

	if strings.Contains(errStr, "timeout") {
		return common.WriteError(cmd, "timeout", "request timed out")
	}
	if strings.Contains(errStr, "connection") || strings.Contains(errStr, "refused") {
		return common.WriteError(cmd, "connection_error", "cannot connect to Ark API")
	}
	if strings.Contains(errStr, "no such host") || strings.Contains(errStr, "dns") {
		return common.WriteError(cmd, "connection_error", "cannot resolve Ark API host")
	}

	return common.WriteError(cmd, "api_error", err.Error())
}

func handleVideoHTTPError(cmd *cobra.Command, statusCode int, body string) error {
	switch statusCode {
	case http.StatusUnauthorized:
		return common.WriteError(cmd, "invalid_api_key", "API key is invalid or revoked")
	case http.StatusForbidden:
		return common.WriteError(cmd, "permission_denied", "API key lacks required permissions")
	case http.StatusNotFound:
		return common.WriteError(cmd, "task_not_found", "task not found")
	case http.StatusTooManyRequests:
		if strings.Contains(body, "quota") {
			return common.WriteError(cmd, "quota_exceeded", "API quota exhausted")
		}
		return common.WriteError(cmd, "rate_limit", "too many requests")
	case http.StatusBadRequest:
		if strings.Contains(body, "safety") || strings.Contains(body, "policy") {
			return common.WriteError(cmd, "content_policy", "content violates safety policy")
		}
		return common.WriteError(cmd, "invalid_request", "invalid request parameters")
	case http.StatusInternalServerError:
		return common.WriteError(cmd, "server_error", "Ark server error")
	case http.StatusServiceUnavailable:
		return common.WriteError(cmd, "server_overloaded", "Ark server overloaded")
	}

	return common.WriteError(cmd, "api_error", fmt.Sprintf("HTTP %d: %s", statusCode, body))
}
