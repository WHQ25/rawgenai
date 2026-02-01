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

type videoDownloadFlags struct {
	output  string
	variant string
}

type videoDownloadResponse struct {
	Success bool   `json:"success"`
	VideoID string `json:"video_id"`
	Variant string `json:"variant"`
	File    string `json:"file"`
}

var videoDownloadCmd = newVideoDownloadCmd()

func newVideoDownloadCmd() *cobra.Command {
	flags := &videoDownloadFlags{}

	cmd := &cobra.Command{
		Use:           "download <video_id>",
		Short:         "Download video content",
		Long:          "Download video, thumbnail, or spritesheet from a completed video job.",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVideoDownload(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.output, "output", "o", "", "Output file path")
	cmd.Flags().StringVar(&flags.variant, "variant", "video", "Content type: video, thumbnail, spritesheet")

	return cmd
}

func runVideoDownload(cmd *cobra.Command, args []string, flags *videoDownloadFlags) error {
	videoID := strings.TrimSpace(args[0])
	if videoID == "" {
		return writeError(cmd, "missing_video_id", "video_id is required")
	}

	// Validate variant
	if !validVariants[flags.variant] {
		return writeError(cmd, "invalid_variant", "variant must be: video, thumbnail, spritesheet")
	}

	// Validate output
	if flags.output == "" {
		return writeError(cmd, "missing_output", "output file is required, use -o flag")
	}

	expectedExt := variantExtensions[flags.variant]
	ext := strings.ToLower(filepath.Ext(flags.output))
	if ext != expectedExt {
		return writeError(cmd, "invalid_format", fmt.Sprintf("output file must be %s for variant '%s'", expectedExt, flags.variant))
	}

	// Check API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return writeError(cmd, "missing_api_key", "OPENAI_API_KEY environment variable is not set")
	}

	client := oai.NewClient()
	ctx := context.Background()

	// Get video status first
	video, err := client.Videos.Get(ctx, videoID)
	if err != nil {
		return handleVideoAPIError(cmd, err)
	}

	// Check if video is ready
	if video.Status != oai.VideoStatusCompleted {
		return writeError(cmd, "video_not_ready", fmt.Sprintf("video is not ready for download, current status: %s", video.Status))
	}

	// Download content
	params := oai.VideoDownloadContentParams{
		Variant: oai.VideoDownloadContentParamsVariant(flags.variant),
	}

	resp, err := client.Videos.DownloadContent(ctx, videoID, params)
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

	result := videoDownloadResponse{
		Success: true,
		VideoID: videoID,
		Variant: flags.variant,
		File:    absPath,
	}
	return writeSuccess(cmd, result)
}
