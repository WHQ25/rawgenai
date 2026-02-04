package image

import (
	"strings"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/cli/hunyuan/shared"
	"github.com/spf13/cobra"
	aiart "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/aiart/v20221229"
	tccommon "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
)

// statusCodeMap maps API status codes to human-readable status strings.
var statusCodeMap = map[string]string{
	"1": "waiting",
	"2": "running",
	"4": "failed",
	"5": "done",
}

type statusFlags struct {
	verbose bool
	region  string
}

func newStatusCmd() *cobra.Command {
	flags := &statusFlags{}

	cmd := &cobra.Command{
		Use:           "status <job_id>",
		Short:         "Get image generation status",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd, args, flags)
		},
	}

	cmd.Flags().BoolVarP(&flags.verbose, "verbose", "v", false, "Show image URLs")
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
	client, err := shared.NewAiartClient(secretID, secretKey, flags.region)
	if err != nil {
		return common.WriteError(cmd, "api_error", "failed to create SDK client: "+err.Error())
	}

	// Build request
	req := aiart.NewQueryTextToImageJobRequest()
	req.JobId = tccommon.StringPtr(jobID)

	// Send request
	resp, err := client.QueryTextToImageJob(req)
	if err != nil {
		return shared.HandleSDKError(cmd, err)
	}

	if resp.Response == nil {
		return common.WriteError(cmd, "response_error", "no data in response")
	}

	r := resp.Response

	// Check for failure
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

	// Build output
	status := statusCodeMap[statusCode]
	if status == "" {
		status = statusCode
	}

	output := map[string]any{
		"success": true,
		"job_id":  jobID,
		"status":  status,
	}

	if r.ResultImage != nil && len(r.ResultImage) > 0 {
		output["image_count"] = len(r.ResultImage)
		if flags.verbose {
			urls := make([]string, 0, len(r.ResultImage))
			for _, u := range r.ResultImage {
				if u != nil {
					urls = append(urls, *u)
				}
			}
			output["images"] = urls
		}
	}

	if r.RevisedPrompt != nil && len(r.RevisedPrompt) > 0 && flags.verbose {
		prompts := make([]string, 0, len(r.RevisedPrompt))
		for _, p := range r.RevisedPrompt {
			if p != nil {
				prompts = append(prompts, *p)
			}
		}
		output["revised_prompts"] = prompts
	}

	return common.WriteSuccess(cmd, output)
}
