package video

import (
	"context"
	"fmt"
	"strings"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
	"google.golang.org/genai"
)

type extendFlags struct {
	promptFile string
	model      string
	negative   string
}

type extendResponse struct {
	Success     bool   `json:"success"`
	OperationID string `json:"operation_id"`
	Status      string `json:"status"`
	Model       string `json:"model"`
}

var extendCmd = newExtendCmd()

func newExtendCmd() *cobra.Command {
	flags := &extendFlags{}

	cmd := &cobra.Command{
		Use:   "extend <operation_id> [prompt]",
		Short: "Extend a previously generated video",
		Long: `Extend a previously generated Veo video by ~7 seconds.

The source video must be from a previous Veo generation (within 2 days).
Extension is limited to 720p resolution and can be done up to 20 times.

IMPORTANT: Save the new operation_id from the response. The extended video
is a new generation and will also be deleted after 2 days.`,
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExtend(cmd, args, flags)
		},
	}

	cmd.Flags().StringVar(&flags.promptFile, "prompt-file", "", "Input prompt file")
	cmd.Flags().StringVarP(&flags.model, "model", "m", "veo-3.1", "Model: veo-3.1, veo-3.1-fast")
	cmd.Flags().StringVar(&flags.negative, "negative", "", "Negative prompt (what to avoid)")

	return cmd
}

func runExtend(cmd *cobra.Command, args []string, flags *extendFlags) error {
	// Get operation ID
	operationID := strings.TrimSpace(args[0])
	if operationID == "" {
		return common.WriteError(cmd, "missing_operation_id", "operation_id is required")
	}

	// Get prompt from remaining args, file, or stdin
	promptArgs := args[1:]
	prompt, err := getPrompt(promptArgs, flags.promptFile, cmd.InOrStdin())
	if err != nil {
		return common.WriteError(cmd, "missing_prompt", "prompt is required for video extension")
	}

	// Validate model
	modelID, ok := modelIDs[flags.model]
	if !ok {
		return common.WriteError(cmd, "invalid_model", fmt.Sprintf("invalid model '%s', use 'veo-3.1' or 'veo-3.1-fast'", flags.model))
	}

	// Check API key
	apiKey := config.GetAPIKey("GEMINI_API_KEY", "GOOGLE_API_KEY")
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("GEMINI_API_KEY", "GOOGLE_API_KEY"))
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

	// Get the previous operation to retrieve the video
	prevOp := &genai.GenerateVideosOperation{Name: operationID}
	prevOp, err = client.Operations.GetVideosOperation(ctx, prevOp, nil)
	if err != nil {
		return handleAPIError(cmd, err)
	}

	// Check if previous operation is completed
	if !prevOp.Done {
		return common.WriteError(cmd, "video_not_ready", "source video generation is not completed yet")
	}

	// Check if previous operation has video
	if prevOp.Response == nil || len(prevOp.Response.GeneratedVideos) == 0 {
		return common.WriteError(cmd, "no_video", "no video found in the source operation")
	}

	// Get the video from previous operation
	sourceVideo := prevOp.Response.GeneratedVideos[0].Video
	if sourceVideo == nil {
		return common.WriteError(cmd, "no_video", "no video found in the source operation")
	}

	// Download video content if needed
	if len(sourceVideo.VideoBytes) == 0 && sourceVideo.URI != "" {
		_, err = client.Files.Download(ctx, sourceVideo, nil)
		if err != nil {
			return common.WriteError(cmd, "download_error", fmt.Sprintf("failed to download source video: %s", err.Error()))
		}
	}

	// Build source for extension
	source := &genai.GenerateVideosSource{
		Prompt: prompt,
		Video:  sourceVideo,
	}

	// Build config - extension requires 720p and allow_all for personGeneration
	config := &genai.GenerateVideosConfig{
		NumberOfVideos:   1,
		Resolution:       "720p",
		PersonGeneration: "allow_all",
	}

	if flags.negative != "" {
		config.NegativePrompt = flags.negative
	}

	// Call API
	op, err := client.Models.GenerateVideosFromSource(ctx, modelID, source, config)
	if err != nil {
		return handleAPIError(cmd, err)
	}

	// Determine status
	status := "running"
	if op.Done {
		status = "completed"
	}

	result := extendResponse{
		Success:     true,
		OperationID: op.Name,
		Status:      status,
		Model:       modelID,
	}

	return common.WriteSuccess(cmd, result)
}
