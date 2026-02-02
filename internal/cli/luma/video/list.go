package video

import (
	"encoding/json"
	"fmt"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/cli/luma/shared"
	"github.com/spf13/cobra"
)

type listFlags struct {
	limit  int
	offset int
}

func newListCmd() *cobra.Command {
	flags := &listFlags{}

	cmd := &cobra.Command{
		Use:           "list",
		Short:         "List generations",
		Long:          "List video generation tasks with pagination.",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd, flags)
		},
	}

	cmd.Flags().IntVar(&flags.limit, "limit", 10, "Number of results to return (1-100)")
	cmd.Flags().IntVar(&flags.offset, "offset", 0, "Offset for pagination")

	return cmd
}

func runList(cmd *cobra.Command, flags *listFlags) error {
	// Validate limit
	if flags.limit < 1 || flags.limit > 100 {
		return common.WriteError(cmd, "invalid_limit", "limit must be between 1 and 100")
	}

	// Validate offset
	if flags.offset < 0 {
		return common.WriteError(cmd, "invalid_offset", "offset must be non-negative")
	}

	// Check API key
	if shared.GetLumaAPIKey() == "" {
		return common.WriteError(cmd, "missing_api_key",
			"LUMA_API_KEY not found. Set it with: rawgenai config set luma_api_key <your-key>")
	}

	// Build endpoint with query params
	endpoint := fmt.Sprintf("/generations?limit=%d&offset=%d", flags.limit, flags.offset)

	req, err := shared.CreateRequest("GET", endpoint, nil)
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

	var listResp shared.ListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return common.WriteError(cmd, "decode_error", err.Error())
	}

	// Build output
	generations := make([]map[string]interface{}, 0, len(listResp.Generations))
	for _, gen := range listResp.Generations {
		item := map[string]interface{}{
			"task_id":    gen.ID,
			"state":      gen.State,
			"created_at": gen.CreatedAt,
		}
		if gen.GenerationType != "" {
			item["generation_type"] = gen.GenerationType
		}
		if gen.Model != "" {
			item["model"] = gen.Model
		}
		if gen.FailureReason != "" {
			item["failure_reason"] = gen.FailureReason
		}
		generations = append(generations, item)
	}

	return common.WriteSuccess(cmd, map[string]interface{}{
		"generations": generations,
		"count":       listResp.Count,
		"limit":       listResp.Limit,
		"offset":      listResp.Offset,
		"has_more":    listResp.HasMore,
	})
}
