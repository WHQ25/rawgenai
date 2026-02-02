package image

import (
	"encoding/json"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/cli/luma/shared"
	"github.com/spf13/cobra"
)

type statusFlags struct {
	verbose bool
}

func newStatusCmd() *cobra.Command {
	flags := &statusFlags{}

	cmd := &cobra.Command{
		Use:   "status <task_id>",
		Short: "Get generation status",
		Long:  "Query the status of an image generation task.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd, args, flags)
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.Flags().BoolVarP(&flags.verbose, "verbose", "v", false, "Show full output including URLs")

	return cmd
}

func runStatus(cmd *cobra.Command, args []string, flags *statusFlags) error {
	taskID := args[0]
	if taskID == "" {
		return common.WriteError(cmd, "missing_task_id", "task_id is required")
	}

	// Check API key
	if shared.GetLumaAPIKey() == "" {
		return common.WriteError(cmd, "missing_api_key",
			"LUMA_API_KEY not found. Set it with: rawgenai config set luma_api_key <your-key>")
	}

	req, err := shared.CreateRequest("GET", "/generations/"+taskID, nil)
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

	var gen shared.Generation
	if err := json.NewDecoder(resp.Body).Decode(&gen); err != nil {
		return common.WriteError(cmd, "decode_error", err.Error())
	}

	result := map[string]interface{}{
		"task_id":         gen.ID,
		"state":           gen.State,
		"generation_type": gen.GenerationType,
		"model":           gen.Model,
		"created_at":      gen.CreatedAt,
	}

	if gen.FailureReason != "" {
		result["failure_reason"] = gen.FailureReason
	}

	// Include asset URLs in verbose mode
	if flags.verbose && gen.Assets != nil {
		assets := make(map[string]string)
		if gen.Assets.Image != "" {
			assets["image"] = gen.Assets.Image
		}
		if len(assets) > 0 {
			result["assets"] = assets
		}
	}

	return common.WriteSuccess(cmd, result)
}
