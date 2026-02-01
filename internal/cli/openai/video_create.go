package openai

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	oai "github.com/openai/openai-go/v3"
	"github.com/spf13/cobra"
)

var validSizes = map[string]bool{
	"1280x720":  true,
	"720x1280":  true,
	"1792x1024": true,
	"1024x1792": true,
}

var validDurations = map[int]bool{
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

type videoCreateFlags struct {
	file     string
	image    string
	model    string
	size     string
	duration int
}

type videoCreateResponse struct {
	Success   bool   `json:"success"`
	VideoID   string `json:"video_id"`
	Status    string `json:"status"`
	Model     string `json:"model"`
	Size      string `json:"size"`
	Duration  int    `json:"duration"`
	CreatedAt int64  `json:"created_at"`
}

var videoCreateCmd = newVideoCreateCmd()

func newVideoCreateCmd() *cobra.Command {
	flags := &videoCreateFlags{}

	cmd := &cobra.Command{
		Use:           "create [prompt]",
		Short:         "Create a video generation job",
		Long:          "Create a video generation job and return the video ID. Use 'video status' to check progress and 'video download' to retrieve the result.",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVideoCreate(cmd, args, flags)
		},
	}

	cmd.Flags().StringVar(&flags.file, "file", "", "Input prompt file")
	cmd.Flags().StringVarP(&flags.image, "image", "i", "", "First frame image (JPEG/PNG/WebP)")
	cmd.Flags().StringVarP(&flags.model, "model", "m", "sora-2", "Model name (sora-2, sora-2-pro)")
	cmd.Flags().StringVarP(&flags.size, "size", "s", "1280x720", "Video resolution")
	cmd.Flags().IntVarP(&flags.duration, "duration", "d", 4, "Video duration in seconds (4, 8, 12)")

	return cmd
}

func runVideoCreate(cmd *cobra.Command, args []string, flags *videoCreateFlags) error {
	// Get prompt from args, file, or stdin
	prompt, err := getVideoPrompt(args, flags.file, cmd.InOrStdin())
	if err != nil {
		return writeError(cmd, "missing_prompt", err.Error())
	}

	// Validate size
	if !validSizes[flags.size] {
		return writeError(cmd, "invalid_size", fmt.Sprintf("invalid size '%s', allowed: 1280x720, 720x1280, 1792x1024, 1024x1792", flags.size))
	}

	// Validate duration
	if !validDurations[flags.duration] {
		return writeError(cmd, "invalid_duration", fmt.Sprintf("invalid duration '%d', allowed: 4, 8, 12", flags.duration))
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
		Seconds: oai.VideoSeconds(fmt.Sprintf("%d", flags.duration)),
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

	result := videoCreateResponse{
		Success:   true,
		VideoID:   video.ID,
		Status:    string(video.Status),
		Model:     flags.model,
		Size:      flags.size,
		Duration:  flags.duration,
		CreatedAt: video.CreatedAt,
	}

	return writeSuccess(cmd, result)
}

func getVideoPrompt(args []string, filePath string, stdin io.Reader) (string, error) {
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
