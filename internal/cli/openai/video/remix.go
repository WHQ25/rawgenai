package video

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	oai "github.com/openai/openai-go/v3"
	"github.com/spf13/cobra"
)

type remixFlags struct {
	promptFile string
}

type remixResponse struct {
	Success       bool   `json:"success"`
	VideoID       string `json:"video_id"`
	Status        string `json:"status"`
	RemixedFromID string `json:"remixed_from_id"`
	CreatedAt     int64  `json:"created_at"`
}

var remixCmd = newRemixCmd()

func newRemixCmd() *cobra.Command {
	flags := &remixFlags{}

	cmd := &cobra.Command{
		Use:           "remix <video_id> [prompt]",
		Short:         "Create a remix from an existing video",
		Long:          "Create a new video based on an existing video with a new prompt.",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRemix(cmd, args, flags)
		},
	}

	cmd.Flags().StringVar(&flags.promptFile, "prompt-file", "", "Input prompt file")

	return cmd
}

func runRemix(cmd *cobra.Command, args []string, flags *remixFlags) error {
	videoID := strings.TrimSpace(args[0])
	if videoID == "" {
		return common.WriteError(cmd, "missing_video_id", "video_id is required")
	}

	// Get prompt from remaining args, file, or stdin
	promptArgs := args[1:]
	prompt, err := getRemixPrompt(promptArgs, flags.promptFile, cmd.InOrStdin())
	if err != nil {
		return common.WriteError(cmd, "missing_prompt", err.Error())
	}

	// Check API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", "OPENAI_API_KEY environment variable is not set")
	}

	client := oai.NewClient()
	ctx := context.Background()

	params := oai.VideoRemixParams{
		Prompt: prompt,
	}

	video, err := client.Videos.Remix(ctx, videoID, params)
	if err != nil {
		return handleAPIError(cmd, err)
	}

	result := remixResponse{
		Success:       true,
		VideoID:       video.ID,
		Status:        string(video.Status),
		RemixedFromID: video.RemixedFromVideoID,
		CreatedAt:     video.CreatedAt,
	}

	return common.WriteSuccess(cmd, result)
}

func getRemixPrompt(args []string, filePath string, stdin io.Reader) (string, error) {
	// From positional arguments
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
