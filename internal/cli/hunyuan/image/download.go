package image

import (
	"fmt"
	"strings"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/cli/hunyuan/shared"
	"github.com/spf13/cobra"
	aiart "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/aiart/v20221229"
	tccommon "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
)

type downloadFlags struct {
	output string
	index  int
	region string
}

func newDownloadCmd() *cobra.Command {
	flags := &downloadFlags{}

	cmd := &cobra.Command{
		Use:           "download <job_id_or_url>",
		Short:         "Download a generated image (by job ID or URL)",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDownload(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.output, "output", "o", "", "Output file path (required)")
	cmd.Flags().IntVar(&flags.index, "index", 0, "Image index (0-based)")
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

	// Job ID mode: validate index, query API, then download
	jobID := arg

	if flags.index < 0 {
		return common.WriteError(cmd, "invalid_index", "index must be >= 0")
	}

	// Check credentials
	secretID, secretKey, err := shared.CheckCredentials(cmd)
	if err != nil {
		return err
	}

	// Create SDK client
	client, err := shared.NewAiartClient(secretID, secretKey, flags.region)
	if err != nil {
		return common.WriteError(cmd, "api_error", "failed to create SDK client: "+err.Error())
	}

	// Query job status to get image URL
	req := aiart.NewQueryTextToImageJobRequest()
	req.JobId = tccommon.StringPtr(jobID)

	resp, err := client.QueryTextToImageJob(req)
	if err != nil {
		return shared.HandleSDKError(cmd, err)
	}

	if resp.Response == nil {
		return common.WriteError(cmd, "response_error", "no data in response")
	}

	r := resp.Response

	// Check status
	statusCode := ""
	if r.JobStatusCode != nil {
		statusCode = *r.JobStatusCode
	}

	if statusCode == "4" {
		msg := "image generation failed"
		if r.JobErrorMsg != nil && *r.JobErrorMsg != "" {
			msg = *r.JobErrorMsg
		}
		return common.WriteError(cmd, "image_failed", msg)
	}

	if statusCode != "5" {
		status := statusCodeMap[statusCode]
		if status == "" {
			status = statusCode
		}
		return common.WriteError(cmd, "task_not_done", fmt.Sprintf("task is not finished (status: %s)", status))
	}

	// Get image URL
	if r.ResultImage == nil || len(r.ResultImage) == 0 {
		return common.WriteError(cmd, "no_result", "no images in result")
	}

	if flags.index >= len(r.ResultImage) {
		return common.WriteError(cmd, "invalid_index", fmt.Sprintf("index %d out of range (total: %d)", flags.index, len(r.ResultImage)))
	}

	imageURL := ""
	if r.ResultImage[flags.index] != nil {
		imageURL = *r.ResultImage[flags.index]
	}
	if imageURL == "" {
		return common.WriteError(cmd, "no_result", "image URL not available")
	}

	// Download file
	if err := shared.DownloadFile(cmd, imageURL, flags.output); err != nil {
		return err
	}

	return common.WriteSuccess(cmd, map[string]any{
		"success": true,
		"job_id":  jobID,
		"file":    shared.AbsPath(flags.output),
	})
}
