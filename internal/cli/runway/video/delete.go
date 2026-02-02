package video

import (
	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/cli/runway/shared"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

func newDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "delete <task_id>",
		Short:         "Delete or cancel a task",
		Long:          "Delete a completed task or cancel a running/pending task.",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(cmd, args)
		},
	}

	return cmd
}

func runDelete(cmd *cobra.Command, args []string) error {
	// 1. Validate required: task_id
	if len(args) == 0 {
		return common.WriteError(cmd, "missing_task_id", "task_id is required")
	}
	taskID := args[0]

	// 2. Check API key
	apiKey := shared.GetRunwayAPIKey()
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key",
			config.GetMissingKeyMessage("RUNWAY_API_KEY"))
	}

	// 3. Make API request
	req, err := shared.CreateRequest("DELETE", "/v1/tasks/"+taskID, nil)
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
