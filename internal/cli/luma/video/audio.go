package video

import (
	"bytes"
	"encoding/json"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/cli/luma/shared"
	"github.com/spf13/cobra"
)

type audioFlags struct {
	negativePrompt string
	promptFile     string
}

func newAudioCmd() *cobra.Command {
	flags := &audioFlags{}

	cmd := &cobra.Command{
		Use:   "audio <task_id> [prompt]",
		Short: "Add audio to a video generation",
		Long:  "Add AI-generated audio to an existing video generation.",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAudio(cmd, args, flags)
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.Flags().StringVar(&flags.negativePrompt, "negative-prompt", "", "Negative prompt for audio generation")
	cmd.Flags().StringVarP(&flags.promptFile, "prompt-file", "f", "", "Read prompt from file")

	return cmd
}

func runAudio(cmd *cobra.Command, args []string, flags *audioFlags) error {
	taskID := args[0]
	if taskID == "" {
		return common.WriteError(cmd, "missing_task_id", "task_id is required")
	}

	// Get prompt (optional)
	promptArgs := args[1:]
	prompt, _ := shared.GetPrompt(promptArgs, flags.promptFile, nil)

	// Check API key
	if shared.GetLumaAPIKey() == "" {
		return common.WriteError(cmd, "missing_api_key",
			"LUMA_API_KEY not found. Set it with: rawgenai config set luma_api_key <your-key>")
	}

	// Build request body
	body := map[string]interface{}{
		"generation_type": "add_audio",
	}

	if prompt != "" {
		body["prompt"] = prompt
	}

	if flags.negativePrompt != "" {
		body["negative_prompt"] = flags.negativePrompt
	}

	jsonBody, _ := json.Marshal(body)
	req, err := shared.CreateRequest("POST", "/generations/"+taskID+"/audio", bytes.NewReader(jsonBody))
	if err != nil {
		return common.WriteError(cmd, "request_error", err.Error())
	}

	resp, err := shared.DoRequest(req)
	if err != nil {
		return shared.HandleHTTPError(cmd, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return shared.HandleAPIError(cmd, resp)
	}

	var gen shared.Generation
	if err := json.NewDecoder(resp.Body).Decode(&gen); err != nil {
		return common.WriteError(cmd, "decode_error", err.Error())
	}

	return common.WriteSuccess(cmd, map[string]interface{}{
		"task_id":    gen.ID,
		"state":      gen.State,
		"created_at": gen.CreatedAt,
	})
}
