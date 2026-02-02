package audio

import (
	"bytes"
	"encoding/json"
	"os"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/cli/runway/shared"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

type isolationFlags struct {
	input string
}

func newIsolationCmd() *cobra.Command {
	flags := &isolationFlags{}

	cmd := &cobra.Command{
		Use:           "isolation",
		Short:         "Isolate voice from background",
		Long:          "Isolate the voice from background audio. Duration must be 4.6-3600 seconds.",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runIsolation(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.input, "input", "i", "", "Input audio file (URL or local path)")

	return cmd
}

func runIsolation(cmd *cobra.Command, args []string, flags *isolationFlags) error {
	// 1. Validate required: input
	if flags.input == "" {
		return common.WriteError(cmd, "missing_input", "input file is required (-i)")
	}

	// 2. Validate file existence (local files only)
	if !shared.IsURL(flags.input) {
		if _, err := os.Stat(flags.input); os.IsNotExist(err) {
			return common.WriteError(cmd, "input_not_found", "input file not found: "+flags.input)
		}
	}

	// 3. Check API key
	apiKey := shared.GetRunwayAPIKey()
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key",
			config.GetMissingKeyMessage("RUNWAY_API_KEY"))
	}

	// 4. Resolve input URI
	inputURI, err := shared.ResolveMediaURI(flags.input, "audio")
	if err != nil {
		return common.WriteError(cmd, "input_read_error", "failed to read input: "+err.Error())
	}

	// 5. Build request body
	body := map[string]any{
		"model":    "eleven_voice_isolation",
		"audioUri": inputURI,
	}

	// 6. Make API request
	bodyJSON, _ := json.Marshal(body)
	req, err := shared.CreateRequest("POST", "/v1/voice_isolation", bytes.NewReader(bodyJSON))
	if err != nil {
		return common.WriteError(cmd, "request_error", err.Error())
	}

	resp, err := shared.DoRequest(req)
	if err != nil {
		return shared.HandleHTTPError(cmd, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return shared.HandleAPIError(cmd, resp)
	}

	// 7. Parse response
	var taskResp shared.TaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&taskResp); err != nil {
		return common.WriteError(cmd, "parse_error", "failed to parse response: "+err.Error())
	}

	// 8. Return task ID
	return common.WriteSuccess(cmd, map[string]any{
		"success": true,
		"task_id": taskResp.ID,
	})
}
