package video

import (
	"fmt"
	"strings"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/cli/hunyuan/shared"
	"github.com/spf13/cobra"
	tccommon "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	vclm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vclm/v20240523"
)

type downloadFlags struct {
	output string
	region string
}

func newDownloadCmd() *cobra.Command {
	flags := &downloadFlags{}

	cmd := &cobra.Command{
		Use:           "download <job_id_or_url>",
		Short:         "Download a generated video (by job ID or URL)",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDownload(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.output, "output", "o", "", "Output file path (required)")
	cmd.Flags().StringVar(&flags.region, "region", shared.DefaultRegion, "Tencent Cloud region")

	return cmd
}

func runDownload(cmd *cobra.Command, args []string, flags *downloadFlags) error {
	if len(args) == 0 || strings.TrimSpace(args[0]) == "" {
		return common.WriteError(cmd, "missing_argument", "job ID or URL is required")
	}
	arg := args[0]

	// Validate output
	if strings.TrimSpace(flags.output) == "" {
		return common.WriteError(cmd, "missing_output", "output file path is required (-o)")
	}

	// Direct URL download
	if shared.IsURL(arg) {
		if err := shared.DownloadFile(cmd, arg, flags.output); err != nil {
			return err
		}
		return common.WriteSuccess(cmd, map[string]any{
			"success": true,
			"file":    shared.AbsPath(flags.output),
		})
	}

	// Job ID mode: query API, then download
	jobID := arg

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

	// Query job status to get video URL
	req := vclm.NewDescribeHunyuanToVideoJobRequest()
	req.JobId = tccommon.StringPtr(jobID)

	resp, err := client.DescribeHunyuanToVideoJob(req)
	if err != nil {
		return shared.HandleSDKError(cmd, err)
	}

	if resp.Response == nil {
		return common.WriteError(cmd, "response_error", "no data in response")
	}

	r := resp.Response

	// Check status
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

	if rawStatus != "DONE" {
		status := statusMap[rawStatus]
		if status == "" {
			status = strings.ToLower(rawStatus)
		}
		return common.WriteError(cmd, "task_not_done", fmt.Sprintf("task is not finished (status: %s)", status))
	}

	// Get video URL
	videoURL := ""
	if r.ResultVideoUrl != nil {
		videoURL = *r.ResultVideoUrl
	}
	if videoURL == "" {
		return common.WriteError(cmd, "no_result", "video URL not available")
	}

	// Download file
	if err := shared.DownloadFile(cmd, videoURL, flags.output); err != nil {
		return err
	}

	return common.WriteSuccess(cmd, map[string]any{
		"success": true,
		"job_id":  jobID,
		"file":    shared.AbsPath(flags.output),
	})
}
