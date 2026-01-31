package openai

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	oai "github.com/openai/openai-go/v3"
	"github.com/spf13/cobra"
)

var validSizes = map[string]bool{
	"1280x720":  true,
	"720x1280":  true,
	"1792x1024": true,
	"1024x1792": true,
}

var validSeconds = map[int]bool{
	4:  true,
	8:  true,
	12: true,
}

var validImageFormats = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
}

type videoFlags struct {
	output  string
	file    string
	image   string
	model   string
	size    string
	seconds int
	noWait  bool
	timeout int
}

type videoResponse struct {
	Success bool   `json:"success"`
	File    string `json:"file,omitempty"`
	VideoID string `json:"video_id,omitempty"`
	Status  string `json:"status,omitempty"`
	Model   string `json:"model,omitempty"`
	Size    string `json:"size,omitempty"`
	Seconds int    `json:"seconds,omitempty"`
}

var videoCmd = newVideoCmd()

func newVideoCmd() *cobra.Command {
	flags := &videoFlags{}

	cmd := &cobra.Command{
		Use:           "video [prompt]",
		Short:         "Generate video using OpenAI Sora models",
		Long:          "Generate video from text prompt using OpenAI Sora models (sora-2, sora-2-pro).",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVideo(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.output, "output", "o", "", "Output file path (must be .mp4)")
	cmd.Flags().StringVar(&flags.file, "file", "", "Input prompt file")
	cmd.Flags().StringVarP(&flags.image, "image", "i", "", "First frame image (JPEG/PNG/WebP)")
	cmd.Flags().StringVarP(&flags.model, "model", "m", "sora-2", "Model name")
	cmd.Flags().StringVarP(&flags.size, "size", "s", "1280x720", "Video resolution")
	cmd.Flags().IntVar(&flags.seconds, "seconds", 4, "Video duration (4, 8, 12)")
	cmd.Flags().BoolVar(&flags.noWait, "no-wait", false, "Return immediately with job ID")
	cmd.Flags().IntVar(&flags.timeout, "timeout", 600, "Max wait time in seconds")

	return cmd
}

func runVideo(cmd *cobra.Command, args []string, flags *videoFlags) error {
	// Get prompt from args, file, or stdin
	prompt, err := getPrompt(args, flags.file, cmd.InOrStdin())
	if err != nil {
		return writeError(cmd, "missing_prompt", err.Error())
	}

	// Validate output
	if flags.output == "" {
		return writeError(cmd, "missing_output", "output file is required, use -o flag")
	}

	// Validate format (must be .mp4)
	ext := strings.ToLower(filepath.Ext(flags.output))
	if ext != ".mp4" {
		return writeError(cmd, "invalid_format", "output file must be .mp4")
	}

	// Validate size
	if !validSizes[flags.size] {
		return writeError(cmd, "invalid_size", fmt.Sprintf("invalid size '%s', allowed: 1280x720, 720x1280, 1792x1024, 1024x1792", flags.size))
	}

	// Validate seconds
	if !validSeconds[flags.seconds] {
		return writeError(cmd, "invalid_seconds", fmt.Sprintf("invalid seconds '%d', allowed: 4, 8, 12", flags.seconds))
	}

	// Check API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return writeError(cmd, "missing_api_key", "OPENAI_API_KEY environment variable is not set")
	}

	// Validate image if provided
	if flags.image != "" {
		if _, err := os.Stat(flags.image); os.IsNotExist(err) {
			return writeError(cmd, "image_not_found", fmt.Sprintf("image file not found: %s", flags.image))
		}
		imgExt := strings.ToLower(filepath.Ext(flags.image))
		if !validImageFormats[imgExt] {
			return writeError(cmd, "invalid_image_format", fmt.Sprintf("unsupported image format '%s', supported: jpg, jpeg, png, webp", imgExt))
		}
	}

	// Call OpenAI API
	client := oai.NewClient()
	ctx := context.Background()

	params := oai.VideoNewParams{
		Model:   oai.VideoModel(flags.model),
		Prompt:  prompt,
		Size:    oai.VideoSize(flags.size),
		Seconds: oai.VideoSeconds(fmt.Sprintf("%d", flags.seconds)),
	}

	// Add input reference image if provided
	if flags.image != "" {
		imgFile, err := os.Open(flags.image)
		if err != nil {
			return writeError(cmd, "image_not_found", fmt.Sprintf("cannot open image file: %s", err.Error()))
		}
		defer imgFile.Close()
		params.InputReference = oai.File(imgFile, filepath.Base(flags.image), getImageMimeType(flags.image))
	}

	// Create video job
	video, err := client.Videos.New(ctx, params)
	if err != nil {
		return handleVideoAPIError(cmd, err)
	}

	// If no-wait, return immediately with job ID
	if flags.noWait {
		result := videoResponse{
			Success: true,
			VideoID: video.ID,
			Status:  string(video.Status),
			Model:   flags.model,
			Size:    flags.size,
			Seconds: flags.seconds,
		}
		return writeSuccess(cmd, result)
	}

	// Poll for completion
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(flags.timeout)*time.Second)
	defer cancel()

	for video.Status == oai.VideoStatusQueued || video.Status == oai.VideoStatusInProgress {
		select {
		case <-timeoutCtx.Done():
			return writeError(cmd, "timeout", fmt.Sprintf("video generation did not complete within %d seconds, video_id: %s", flags.timeout, video.ID))
		default:
			time.Sleep(2 * time.Second)
			video, err = client.Videos.Get(ctx, video.ID)
			if err != nil {
				return handleVideoAPIError(cmd, err)
			}
		}
	}

	// Check if generation failed
	if video.Status == oai.VideoStatusFailed {
		errMsg := "video generation failed"
		if video.Error.Message != "" {
			errMsg = video.Error.Message
		}
		return writeError(cmd, "generation_failed", errMsg)
	}

	// Download video content
	resp, err := client.Videos.DownloadContent(ctx, video.ID, oai.VideoDownloadContentParams{})
	if err != nil {
		return handleVideoAPIError(cmd, err)
	}
	defer resp.Body.Close()

	// Get absolute path for output
	absPath, err := filepath.Abs(flags.output)
	if err != nil {
		absPath = flags.output
	}

	// Write to file
	outFile, err := os.Create(absPath)
	if err != nil {
		return writeError(cmd, "output_write_error", fmt.Sprintf("cannot create output file: %s", err.Error()))
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return writeError(cmd, "output_write_error", fmt.Sprintf("cannot write output file: %s", err.Error()))
	}

	// Return success
	result := videoResponse{
		Success: true,
		File:    absPath,
		VideoID: video.ID,
		Model:   flags.model,
		Size:    flags.size,
		Seconds: flags.seconds,
	}
	return writeSuccess(cmd, result)
}

func getPrompt(args []string, filePath string, stdin io.Reader) (string, error) {
	// From positional argument
	if len(args) > 0 {
		prompt := strings.TrimSpace(strings.Join(args, " "))
		if prompt != "" {
			return prompt, nil
		}
	}

	// From file
	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("cannot read file: %w", err)
		}
		prompt := strings.TrimSpace(string(data))
		if prompt != "" {
			return prompt, nil
		}
	}

	// From stdin (only if not a terminal)
	if stdin != nil {
		if f, ok := stdin.(*os.File); ok {
			stat, _ := f.Stat()
			if (stat.Mode() & os.ModeCharDevice) != 0 {
				return "", fmt.Errorf("no prompt provided, use positional argument, --file flag, or pipe from stdin")
			}
		}
		data, err := io.ReadAll(stdin)
		if err != nil {
			return "", fmt.Errorf("cannot read stdin: %w", err)
		}
		prompt := strings.TrimSpace(string(data))
		if prompt != "" {
			return prompt, nil
		}
	}

	return "", fmt.Errorf("no prompt provided, use positional argument, --file flag, or pipe from stdin")
}

func getImageMimeType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
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

	// Network errors
	if strings.Contains(err.Error(), "timeout") {
		return writeError(cmd, "timeout", "Request timed out")
	}
	return writeError(cmd, "connection_error", fmt.Sprintf("Cannot connect to OpenAI API: %s", err.Error()))
}
