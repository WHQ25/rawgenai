package image

import (
	"regexp"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/cli/hunyuan/shared"
	"github.com/spf13/cobra"
	aiart "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/aiart/v20221229"
	tccommon "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
)

var resolutionPattern = regexp.MustCompile(`^\d+:\d+$`)

type createFlags struct {
	images      []string
	resolution  string
	seed        int64
	noRevise    bool
	noWatermark bool
	region      string
	promptFile  string
}

func newCreateCmd() *cobra.Command {
	flags := &createFlags{}

	cmd := &cobra.Command{
		Use:           "create [prompt]",
		Short:         "Create an image generation task",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(cmd, args, flags)
		},
	}

	cmd.Flags().StringArrayVarP(&flags.images, "image", "i", nil, "Reference image URL (repeatable, max 3)")
	cmd.Flags().StringVarP(&flags.resolution, "resolution", "r", "1024:1024", "Resolution W:H, e.g. 1024:1024, 1024:768, 768:1024 (each side 512-2048, area ≤1024x1024)")
	cmd.Flags().Int64Var(&flags.seed, "seed", 0, "Random seed (0=random)")
	cmd.Flags().BoolVar(&flags.noRevise, "no-revise", false, "Disable prompt expansion")
	cmd.Flags().BoolVar(&flags.noWatermark, "no-watermark", false, "Disable watermark")
	cmd.Flags().StringVar(&flags.region, "region", shared.DefaultRegion, "Tencent Cloud region")
	cmd.Flags().StringVarP(&flags.promptFile, "prompt-file", "f", "", "Read prompt from file")

	return cmd
}

func runCreate(cmd *cobra.Command, args []string, flags *createFlags) error {
	// Get prompt
	prompt, err := shared.GetPrompt(args, flags.promptFile, cmd.InOrStdin())
	if err != nil {
		return common.WriteError(cmd, "missing_prompt", err.Error())
	}

	// Validate resolution format
	if !resolutionPattern.MatchString(flags.resolution) {
		return common.WriteError(cmd, "invalid_resolution", "resolution must be in W:H format (e.g. 1024:1024)")
	}

	// Validate image count
	if len(flags.images) > 3 {
		return common.WriteError(cmd, "invalid_image_count", "maximum 3 reference images allowed")
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

	// Build request
	req := aiart.NewSubmitTextToImageJobRequest()
	req.Prompt = tccommon.StringPtr(prompt)
	req.Resolution = tccommon.StringPtr(flags.resolution)

	if len(flags.images) > 0 {
		imgs := make([]*string, len(flags.images))
		for i, img := range flags.images {
			imgs[i] = tccommon.StringPtr(img)
		}
		req.Images = imgs
	}

	if flags.seed != 0 {
		req.Seed = tccommon.Int64Ptr(flags.seed)
	}

	if flags.noRevise {
		req.Revise = tccommon.Int64Ptr(0)
	}

	if flags.noWatermark {
		req.LogoAdd = tccommon.Int64Ptr(0)
	}

	// Send request
	resp, err := client.SubmitTextToImageJob(req)
	if err != nil {
		return shared.HandleSDKError(cmd, err)
	}

	if resp.Response == nil || resp.Response.JobId == nil {
		return common.WriteError(cmd, "response_error", "no job ID in response")
	}

	return common.WriteSuccess(cmd, map[string]any{
		"success": true,
		"job_id":  *resp.Response.JobId,
	})
}
