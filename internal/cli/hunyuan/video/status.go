package video

import (
	"strings"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/cli/hunyuan/shared"
	"github.com/spf13/cobra"
	tccommon "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	vclm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vclm/v20240523"
)

// statusMap maps API status strings to lowercase.
var statusMap = map[string]string{
	"WAIT": "waiting",
	"RUN":  "running",
	"FAIL": "failed",
	"DONE": "done",
}

type statusFlags struct {
	verbose bool
	region  string
}

func newStatusCmd() *cobra.Command {
	flags := &statusFlags{}

	cmd := &cobra.Command{
		Use:           "status <job_id>",
		Short:         "Get video generation status",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd, args, flags)
		},
	}

	cmd.Flags().BoolVarP(&flags.verbose, "verbose", "v", false, "Show video URL")
	cmd.Flags().StringVar(&flags.region, "region", shared.DefaultRegion, "Tencent Cloud region")

	return cmd
}

func runStatus(cmd *cobra.Command, args []string, flags *statusFlags) error {
	// Validate job ID
	if len(args) == 0 || strings.TrimSpace(args[0]) == "" {
		return common.WriteError(cmd, "missing_job_id", "job ID is required")
	}
	jobID := args[0]

	// Check credentials
	secretID, secretKey, err := shared.CheckCredentials(cmd)
	if err != nil {
		return err
	}

	// Create SDK client
	client, err := shared.NewVclmClient(secretID, secretKey, flags.region)
	if err != nil {
		return common.WriteError(cmd, "api_error", "failed to create SDK client: "+err.Error())
	}

	// Build request
	req := vclm.NewDescribeHunyuanToVideoJobRequest()
	req.JobId = tccommon.StringPtr(jobID)

	// Send request
	resp, err := client.DescribeHunyuanToVideoJob(req)
	if err != nil {
		return shared.HandleSDKError(cmd, err)
	}

	if resp.Response == nil {
		return common.WriteError(cmd, "response_error", "no data in response")
	}

	r := resp.Response

	// Get status
	rawStatus := ""
	if r.Status != nil {
		rawStatus = *r.Status
	}

	if rawStatus == "FAIL" {
		msg := "video generation failed"
		if r.ErrorMessage != nil && *r.ErrorMessage != "" {
			msg = *r.ErrorMessage
		}
		return common.WriteError(cmd, "video_failed", msg)
	}

	status := statusMap[rawStatus]
	if status == "" {
		status = strings.ToLower(rawStatus)
	}

	output := map[string]any{
		"success": true,
		"job_id":  jobID,
		"status":  status,
	}

	if r.ResultVideoUrl != nil && *r.ResultVideoUrl != "" {
		if flags.verbose {
			output["video_url"] = *r.ResultVideoUrl
		}
	}

	return common.WriteSuccess(cmd, output)
}
