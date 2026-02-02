package voice

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/cli/minimax/shared"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

type listFlags struct {
	voiceType string
	pretty    bool
}

var validVoiceTypes = map[string]bool{
	"all":              true,
	"system":           true,
	"voice_cloning":    true,
	"voice_generation": true,
}

func newListCmd() *cobra.Command {
	flags := &listFlags{}

	cmd := &cobra.Command{
		Use:           "list",
		Short:         "List available voices",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.voiceType, "type", "t", "all", "Voice type: all, system, voice_cloning, voice_generation")
	cmd.Flags().BoolVar(&flags.pretty, "pretty", false, "Pretty print JSON output")
	return cmd
}

type voiceItem struct {
	VoiceID     string   `json:"voice_id,omitempty"`
	VoiceName   string   `json:"voice_name,omitempty"`
	Description []string `json:"description,omitempty"`
	Language    string   `json:"language,omitempty"`
	Gender      string   `json:"gender,omitempty"`
	CreatedTime string   `json:"created_time,omitempty"`
}

type listResponse struct {
	Success         bool        `json:"success"`
	SystemVoices    []voiceItem `json:"system_voices,omitempty"`
	CloningVoices   []voiceItem `json:"cloning_voices,omitempty"`
	GeneratedVoices []voiceItem `json:"generated_voices,omitempty"`
}

func runList(cmd *cobra.Command, flags *listFlags) error {
	if !validVoiceTypes[flags.voiceType] {
		return common.WriteError(cmd, "invalid_type", "type must be all, system, voice_cloning, or voice_generation")
	}

	apiKey := shared.GetMinimaxAPIKey()
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("MINIMAX_API_KEY"))
	}

	body := map[string]any{
		"voice_type": flags.voiceType,
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot serialize request: %s", err.Error()))
	}

	req, err := shared.CreateRequest("POST", "/v1/get_voice", bytes.NewReader(jsonBody))
	if err != nil {
		return common.WriteError(cmd, "request_error", err.Error())
	}

	resp, err := shared.DoRequest(req)
	if err != nil {
		return common.WriteError(cmd, "request_error", err.Error())
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot read response: %s", err.Error()))
	}

	if resp.StatusCode != http.StatusOK {
		return common.WriteError(cmd, "api_error", fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(respBody)))
	}

	var apiResp struct {
		SystemVoice     []voiceItem `json:"system_voice"`
		VoiceCloning    []voiceItem `json:"voice_cloning"`
		VoiceGeneration []voiceItem `json:"voice_generation"`
		BaseResp        struct {
			StatusCode int    `json:"status_code"`
			StatusMsg  string `json:"status_msg"`
		} `json:"base_resp"`
	}
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot parse response: %s", err.Error()))
	}

	if apiResp.BaseResp.StatusCode != 0 {
		return common.WriteError(cmd, "api_error", fmt.Sprintf("api error %d: %s", apiResp.BaseResp.StatusCode, apiResp.BaseResp.StatusMsg))
	}

	result := listResponse{
		Success:         true,
		SystemVoices:    apiResp.SystemVoice,
		CloningVoices:   apiResp.VoiceCloning,
		GeneratedVoices: apiResp.VoiceGeneration,
	}

	if flags.pretty {
		output, _ := json.MarshalIndent(result, "", "  ")
		fmt.Fprintln(cmd.OutOrStdout(), string(output))
		return nil
	}
	return common.WriteSuccess(cmd, result)
}
