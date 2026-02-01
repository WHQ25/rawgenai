package video

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/spf13/cobra"
	"google.golang.org/genai"
)

// Model name mapping (veo-3.1 only)
var modelIDs = map[string]string{
	"veo-3.1":      "veo-3.1-generate-preview",
	"veo-3.1-fast": "veo-3.1-fast-generate-preview",
}

// Valid aspect ratios for Veo
var validAspects = map[string]bool{
	"16:9": true,
	"9:16": true,
}

// Valid resolutions for Veo
var validResolutions = map[string]bool{
	"720p":  true,
	"1080p": true,
	"4k":    true,
}

// Valid durations for Veo (in seconds)
var validDurations = map[int]bool{
	4: true,
	6: true,
	8: true,
}

// Valid image formats for first frame
var validImageFormats = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
}

type createFlags struct {
	promptFile string
	firstFrame string
	lastFrame  string
	ref        []string
	model      string
	aspect     string
	resolution string
	duration   int
	negative   string
	seed       int
}

type createResponse struct {
	Success     bool   `json:"success"`
	OperationID string `json:"operation_id"`
	Status      string `json:"status"`
	Model       string `json:"model"`
	Aspect      string `json:"aspect"`
	Resolution  string `json:"resolution"`
	Duration    int    `json:"duration"`
}

var createCmd = newCreateCmd()

func newCreateCmd() *cobra.Command {
	flags := &createFlags{}

	cmd := &cobra.Command{
		Use:   "create [prompt]",
		Short: "Create a video generation job",
		Long: `Create a video generation job and return the operation ID.

Use 'video status' to check progress and 'video download' to retrieve the result.

IMPORTANT: Save the operation_id from the response. You need it to check status,
download the video, and extend. Videos are deleted after 2 days.`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(cmd, args, flags)
		},
	}

	cmd.Flags().StringVar(&flags.promptFile, "prompt-file", "", "Input prompt file")
	cmd.Flags().StringVar(&flags.firstFrame, "first-frame", "", "First frame image (JPEG/PNG)")
	cmd.Flags().StringVar(&flags.lastFrame, "last-frame", "", "Last frame image (JPEG/PNG), requires --first-frame")
	cmd.Flags().StringArrayVar(&flags.ref, "ref", nil, "Reference image (max 3, repeatable)")
	cmd.Flags().StringVarP(&flags.model, "model", "m", "veo-3.1", "Model: veo-3.1, veo-3.1-fast")
	cmd.Flags().StringVarP(&flags.aspect, "aspect", "a", "16:9", "Aspect ratio: 16:9, 9:16")
	cmd.Flags().StringVarP(&flags.resolution, "resolution", "r", "720p", "Resolution: 720p, 1080p, 4k")
	cmd.Flags().IntVarP(&flags.duration, "duration", "d", 8, "Duration in seconds: 4, 6, 8")
	cmd.Flags().StringVar(&flags.negative, "negative", "", "Negative prompt (what to avoid)")
	cmd.Flags().IntVar(&flags.seed, "seed", 0, "Seed for reproducibility")

	return cmd
}

func runCreate(cmd *cobra.Command, args []string, flags *createFlags) error {
	// Get prompt from args, file, or stdin
	prompt, err := getPrompt(args, flags.promptFile, cmd.InOrStdin())
	if err != nil {
		return common.WriteError(cmd, "missing_prompt", err.Error())
	}

	// Validate model
	modelID, ok := modelIDs[flags.model]
	if !ok {
		return common.WriteError(cmd, "invalid_model", fmt.Sprintf("invalid model '%s', use 'veo-3.1' or 'veo-3.1-fast'", flags.model))
	}

	// Validate aspect ratio
	if !validAspects[flags.aspect] {
		return common.WriteError(cmd, "invalid_aspect", fmt.Sprintf("invalid aspect ratio '%s', use '16:9' or '9:16'", flags.aspect))
	}

	// Validate resolution
	if !validResolutions[flags.resolution] {
		return common.WriteError(cmd, "invalid_resolution", fmt.Sprintf("invalid resolution '%s', use '720p', '1080p', or '4k'", flags.resolution))
	}

	// Validate duration
	if !validDurations[flags.duration] {
		return common.WriteError(cmd, "invalid_duration", fmt.Sprintf("invalid duration '%d', use 4, 6, or 8", flags.duration))
	}

	// Validate resolution+duration: 1080p and 4k only support 8s
	if (flags.resolution == "1080p" || flags.resolution == "4k") && flags.duration != 8 {
		return common.WriteError(cmd, "invalid_resolution_duration", fmt.Sprintf("%s resolution only supports 8 second duration", flags.resolution))
	}

	// Check API key
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("GOOGLE_API_KEY")
	}
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", "GEMINI_API_KEY or GOOGLE_API_KEY environment variable is not set")
	}

	// Validate last-frame requires first-frame
	if flags.lastFrame != "" && flags.firstFrame == "" {
		return common.WriteError(cmd, "last_frame_requires_first", "--last-frame requires --first-frame")
	}

	// Check if using reference images
	hasRefImages := len(flags.ref) > 0
	hasFrameImages := flags.firstFrame != "" || flags.lastFrame != ""

	// Validate mutual exclusivity: reference images vs frame images
	if hasRefImages && hasFrameImages {
		return common.WriteError(cmd, "conflicting_image_options", "--ref cannot be used with --first-frame/--last-frame")
	}

	// Validate ref count (max 3)
	if len(flags.ref) > 3 {
		return common.WriteError(cmd, "too_many_refs", "maximum 3 --ref images allowed")
	}

	// Validate first frame if provided
	var firstFrame *genai.Image
	if flags.firstFrame != "" {
		if _, err := os.Stat(flags.firstFrame); os.IsNotExist(err) {
			return common.WriteError(cmd, "first_frame_not_found", fmt.Sprintf("first frame file not found: %s", flags.firstFrame))
		}
		imgExt := strings.ToLower(filepath.Ext(flags.firstFrame))
		if !validImageFormats[imgExt] {
			return common.WriteError(cmd, "invalid_image_format", fmt.Sprintf("unsupported image format '%s', supported: jpg, jpeg, png", imgExt))
		}
		imageData, err := os.ReadFile(flags.firstFrame)
		if err != nil {
			return common.WriteError(cmd, "first_frame_not_found", fmt.Sprintf("cannot read first frame file: %s", err.Error()))
		}
		firstFrame = &genai.Image{
			ImageBytes: imageData,
			MIMEType:   getImageMimeType(flags.firstFrame),
		}
	}

	// Validate last frame if provided
	var lastFrame *genai.Image
	if flags.lastFrame != "" {
		if _, err := os.Stat(flags.lastFrame); os.IsNotExist(err) {
			return common.WriteError(cmd, "last_frame_not_found", fmt.Sprintf("last frame file not found: %s", flags.lastFrame))
		}
		imgExt := strings.ToLower(filepath.Ext(flags.lastFrame))
		if !validImageFormats[imgExt] {
			return common.WriteError(cmd, "invalid_image_format", fmt.Sprintf("unsupported image format '%s', supported: jpg, jpeg, png", imgExt))
		}
		imageData, err := os.ReadFile(flags.lastFrame)
		if err != nil {
			return common.WriteError(cmd, "last_frame_not_found", fmt.Sprintf("cannot read last frame file: %s", err.Error()))
		}
		lastFrame = &genai.Image{
			ImageBytes: imageData,
			MIMEType:   getImageMimeType(flags.lastFrame),
		}
	}

	// Validate and load reference images
	var refImages []*genai.VideoGenerationReferenceImage
	for _, refPath := range flags.ref {
		if _, err := os.Stat(refPath); os.IsNotExist(err) {
			return common.WriteError(cmd, "ref_not_found", fmt.Sprintf("reference image not found: %s", refPath))
		}
		imgExt := strings.ToLower(filepath.Ext(refPath))
		if !validImageFormats[imgExt] {
			return common.WriteError(cmd, "invalid_image_format", fmt.Sprintf("unsupported image format '%s', supported: jpg, jpeg, png", imgExt))
		}
		imageData, err := os.ReadFile(refPath)
		if err != nil {
			return common.WriteError(cmd, "ref_not_found", fmt.Sprintf("cannot read reference image: %s", err.Error()))
		}
		refImages = append(refImages, &genai.VideoGenerationReferenceImage{
			Image: &genai.Image{
				ImageBytes: imageData,
				MIMEType:   getImageMimeType(refPath),
			},
			ReferenceType: genai.VideoGenerationReferenceTypeAsset,
		})
	}

	// Create client
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return common.WriteError(cmd, "client_error", fmt.Sprintf("failed to create client: %s", err.Error()))
	}

	// Build video generation config
	durationSecs := int32(flags.duration)
	config := &genai.GenerateVideosConfig{
		AspectRatio:      flags.aspect,
		Resolution:       flags.resolution,
		NumberOfVideos:   1,
		DurationSeconds:  &durationSecs,
		PersonGeneration: "allow_adult",
	}

	if flags.negative != "" {
		config.NegativePrompt = flags.negative
	}

	if flags.seed != 0 {
		seedVal := int32(flags.seed)
		config.Seed = &seedVal
	}

	if lastFrame != nil {
		config.LastFrame = lastFrame
	}

	if len(refImages) > 0 {
		config.ReferenceImages = refImages
	}

	// Call API
	op, err := client.Models.GenerateVideos(ctx, modelID, prompt, firstFrame, config)
	if err != nil {
		return handleAPIError(cmd, err)
	}

	// Determine status
	status := "running"
	if op.Done {
		status = "completed"
	}

	result := createResponse{
		Success:     true,
		OperationID: op.Name,
		Status:      status,
		Model:       modelID,
		Aspect:      flags.aspect,
		Resolution:  flags.resolution,
		Duration:    flags.duration,
	}

	return common.WriteSuccess(cmd, result)
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
	default:
		return "application/octet-stream"
	}
}
