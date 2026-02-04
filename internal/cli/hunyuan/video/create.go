package video

import (
	"fmt"
	"os"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/cli/hunyuan/shared"
	"github.com/spf13/cobra"
	tccommon "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	vclm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vclm/v20240523"
)

var validResolutions = map[string]bool{
	"720p": true,
}

type createFlags struct {
	image       string
	resolution  string
	noWatermark bool
	region      string
	promptFile  string
}

func newCreateCmd() *cobra.Command {
	flags := &createFlags{}

	cmd := &cobra.Command{
		Use:           "create [prompt]",
		Short:         "Create a video generation task",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.image, "image", "i", "", "Input image for I2V (local path or URL)")
	cmd.Flags().StringVarP(&flags.resolution, "resolution", "r", "720p", "Video resolution: 720p")
	cmd.Flags().BoolVar(&flags.noWatermark, "no-watermark", false, "Disable watermark")
	cmd.Flags().StringVar(&flags.region, "region", shared.DefaultRegion, "Tencent Cloud region")
	cmd.Flags().StringVarP(&flags.promptFile, "prompt-file", "f", "", "Read prompt from file")

	return cmd
}

func runCreate(cmd *cobra.Command, args []string, flags *createFlags) error {
	// Get prompt (optional when image is provided)
	prompt, promptErr := shared.GetPrompt(args, flags.promptFile, cmd.InOrStdin())

	// Need at least prompt or image
	if promptErr != nil && flags.image == "" {
		return common.WriteError(cmd, "missing_prompt", promptErr.Error())
	}

	// Validate resolution
	if !validResolutions[flags.resolution] {
		return common.WriteError(cmd, "invalid_resolution", fmt.Sprintf("invalid resolution '%s', currently only 720p is supported", flags.resolution))
	}

	// Validate image file existence (skip URLs)
	if flags.image != "" && !shared.IsURL(flags.image) {
		if _, err := os.Stat(flags.image); os.IsNotExist(err) {
			return common.WriteError(cmd, "image_not_found", "image file not found: "+flags.image)
		}
	}

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
	req := vclm.NewSubmitHunyuanToVideoJobRequest()

	if prompt != "" {
		req.Prompt = tccommon.StringPtr(prompt)
	}

	req.Resolution = tccommon.StringPtr(flags.resolution)

	if flags.noWatermark {
		req.LogoAdd = tccommon.Int64Ptr(0)
	}

	// Handle image input
	if flags.image != "" {
		img := &vclm.Image{}
		if shared.IsURL(flags.image) {
			img.Url = tccommon.StringPtr(flags.image)
		} else {
			b64, err := shared.ResolveImageBase64(flags.image)
			if err != nil {
				return common.WriteError(cmd, "image_read_error", fmt.Sprintf("cannot read image: %s", err.Error()))
			}
			img.Base64 = tccommon.StringPtr(b64)
		}
		req.Image = img
	}

	// Send request
	resp, err := client.SubmitHunyuanToVideoJob(req)
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
