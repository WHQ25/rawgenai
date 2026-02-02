package video

import (
	"encoding/json"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/cli/runway/shared"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

type statusFlags struct {
	verbose bool
}

func newStatusCmd() *cobra.Command {
	flags := &statusFlags{}

	cmd := &cobra.Command{
		Use:           "status <task_id>",
		Short:         "Query video generation task status",
		Long:          "Query the status of a video generation task.",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd, args, flags)
		},
	}

	cmd.Flags().BoolVarP(&flags.verbose, "verbose", "v", false, "Show full output including URLs")

	return cmd
}

func runStatus(cmd *cobra.Command, args []string, flags *statusFlags) error {
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
	req, err := shared.CreateRequest("GET", "/v1/tasks/"+taskID, nil)
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

	// 4. Parse response
	var taskStatus shared.TaskStatus
	if err := json.NewDecoder(resp.Body).Decode(&taskStatus); err != nil {
		return common.WriteError(cmd, "parse_error", "failed to parse response: "+err.Error())
	}

	// 5. Handle failed status
	if taskStatus.Status == shared.StatusFailed {
		msg := taskStatus.Failure
		if msg == "" {
			msg = "video generation failed"
		}
		return common.WriteError(cmd, "video_failed", msg)
	}

	// 6. Build response
	output := map[string]any{
		"success":    true,
		"task_id":    taskStatus.ID,
		"status":     taskStatus.Status,
		"created_at": taskStatus.CreatedAt,
	}

	// Add output URLs if succeeded and verbose
	if taskStatus.Status == shared.StatusSucceeded && len(taskStatus.Output) > 0 {
		if flags.verbose {
			output["output"] = taskStatus.Output
		}
	}

	return common.WriteSuccess(cmd, output)
}
