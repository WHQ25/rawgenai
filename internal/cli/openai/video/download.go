package video

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/config"
	oai "github.com/openai/openai-go/v3"
	"github.com/spf13/cobra"
)

var validVariants = map[string]bool{
	"video":       true,
	"thumbnail":   true,
	"spritesheet": true,
}

var variantExtensions = map[string]string{
	"video":       ".mp4",
	"thumbnail":   ".jpg",
	"spritesheet": ".jpg",
}

type downloadFlags struct {
	output  string
	variant string
}

type downloadResponse struct {
	Success bool   `json:"success"`
	VideoID string `json:"video_id"`
	Variant string `json:"variant"`
	File    string `json:"file"`
}

var downloadCmd = newDownloadCmd()

func newDownloadCmd() *cobra.Command {
	flags := &downloadFlags{}

	cmd := &cobra.Command{
		Use:           "download <video_id>",
		Short:         "Download video content",
		Long:          "Download video, thumbnail, or spritesheet from a completed video job.",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDownload(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.output, "output", "o", "", "Output file path")
	cmd.Flags().StringVar(&flags.variant, "variant", "video", "Content type: video, thumbnail, spritesheet")

	return cmd
}

func runDownload(cmd *cobra.Command, args []string, flags *downloadFlags) error {
	videoID := strings.TrimSpace(args[0])
	if videoID == "" {
		return common.WriteError(cmd, "missing_video_id", "video_id is required")
	}

	// Validate variant
	if !validVariants[flags.variant] {
		return common.WriteError(cmd, "invalid_variant", "variant must be: video, thumbnail, spritesheet")
	}

	// Validate output
	if flags.output == "" {
		return common.WriteError(cmd, "missing_output", "output file is required, use -o flag")
	}

	expectedExt := variantExtensions[flags.variant]
	ext := strings.ToLower(filepath.Ext(flags.output))
	if ext != expectedExt {
		return common.WriteError(cmd, "invalid_format", fmt.Sprintf("output file must be %s for variant '%s'", expectedExt, flags.variant))
	}

	// Check API key
	apiKey := config.GetAPIKey("OPENAI_API_KEY")
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("OPENAI_API_KEY"))
	}

	client := oai.NewClient()
	ctx := context.Background()

	// Get video status first
	video, err := client.Videos.Get(ctx, videoID)
	if err != nil {
		return handleAPIError(cmd, err)
	}

	// Check if video is ready
	if video.Status != oai.VideoStatusCompleted {
		return common.WriteError(cmd, "video_not_ready", fmt.Sprintf("video is not ready for download, current status: %s", video.Status))
	}

	// Download content
	params := oai.VideoDownloadContentParams{
		Variant: oai.VideoDownloadContentParamsVariant(flags.variant),
	}

	resp, err := client.Videos.DownloadContent(ctx, videoID, params)
	if err != nil {
		return handleAPIError(cmd, err)
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
		return common.WriteError(cmd, "output_write_error", fmt.Sprintf("cannot create output file: %s", err.Error()))
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return common.WriteError(cmd, "output_write_error", fmt.Sprintf("cannot write output file: %s", err.Error()))
	}

	result := downloadResponse{
		Success: true,
		VideoID: videoID,
		Variant: flags.variant,
		File:    absPath,
	}
	return common.WriteSuccess(cmd, result)
}
