package video

import (
	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/cli/luma/shared"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

func newDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "delete <task_id>",
		Short:         "Delete a generation",
		Long:          "Delete a video generation task.",
		Args:          cobra.ExactArgs(1),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(cmd, args)
		},
	}

	return cmd
}

func runDelete(cmd *cobra.Command, args []string) error {
	taskID := args[0]
	if taskID == "" {
		return common.WriteError(cmd, "missing_task_id", "task_id is required")
	}

	// Check API key
	if shared.GetLumaAPIKey() == "" {
		return common.WriteError(cmd, "missing_api_key",
			config.GetMissingKeyMessage("LUMA_API_KEY"))
	}

	// Make DELETE request
	req, err := shared.CreateRequest("DELETE", "/generations/"+taskID, nil)
	if err != nil {
		return common.WriteError(cmd, "request_error", err.Error())
	}

	resp, err := shared.DoRequest(req)
	if err != nil {
		return shared.HandleHTTPError(cmd, err)
	}
	defer resp.Body.Close()

	// DELETE returns 204 No Content on success
	if resp.StatusCode != 204 && resp.StatusCode != 200 {
		return shared.HandleAPIError(cmd, resp)
	}

	return common.WriteSuccess(cmd, map[string]any{
		"success": true,
		"task_id": taskID,
		"deleted": true,
	})
}
