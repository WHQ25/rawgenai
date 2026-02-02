package image

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/cli/minimax/shared"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

// Cmd is the image subcommand
var Cmd = &cobra.Command{
	Use:   "image",
	Short: "MiniMax image generation commands",
	Long:  "Generate images using MiniMax image generation API.",
}

type imageFlags struct {
	output          string
	promptFile      string
	images          []string
	model           string
	aspect          string
	width           int
	height          int
	count           int
	responseFormat  string
	promptOptimizer bool
}

var validImageModels = map[string]bool{
	"image-01":      true,
	"image-01-live": true,
}

var validImageAspects = map[string]bool{
	"1:1":  true,
	"16:9": true,
	"4:3":  true,
	"3:2":  true,
	"2:3":  true,
	"3:4":  true,
	"9:16": true,
	"21:9": true,
}

var validImageFormats = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".webp": true,
}

func init() {
	Cmd.AddCommand(newImageCmd())
}

func newImageCmd() *cobra.Command {
	flags := &imageFlags{}

	cmd := &cobra.Command{
		Use:           "create [prompt]",
		Short:         "Generate images (text-to-image or image-to-image)",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runImage(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.output, "output", "o", "", "Output file path (required)")
	cmd.Flags().StringVar(&flags.promptFile, "prompt-file", "", "Read prompt from file")
	cmd.Flags().StringArrayVarP(&flags.images, "image", "i", nil, "Reference image(s) for i2i (can be repeated)")
	cmd.Flags().StringVarP(&flags.model, "model", "m", "image-01", "Model: image-01, image-01-live (i2i only)")
	cmd.Flags().StringVar(&flags.aspect, "aspect", "1:1", "Aspect ratio: 1:1, 16:9, 4:3, 3:2, 2:3, 3:4, 9:16, 21:9")
	cmd.Flags().IntVar(&flags.width, "width", 0, "Width in pixels (512-2048, multiple of 8; requires --height)")
	cmd.Flags().IntVar(&flags.height, "height", 0, "Height in pixels (512-2048, multiple of 8; requires --width)")
	cmd.Flags().IntVarP(&flags.count, "n", "n", 1, "Number of images to generate (1-9)")
	cmd.Flags().StringVar(&flags.responseFormat, "response-format", "base64", "Response format: base64 or url")
	cmd.Flags().BoolVar(&flags.promptOptimizer, "prompt-optimizer", false, "Enable prompt optimization")

	return cmd
}

type imageResponse struct {
	Success bool     `json:"success"`
	Files   []string `json:"files,omitempty"`
	File    string   `json:"file,omitempty"`
	URLs    []string `json:"urls,omitempty"`
	URL     string   `json:"url,omitempty"`
	Model   string   `json:"model,omitempty"`
	Count   int      `json:"count,omitempty"`
}

func runImage(cmd *cobra.Command, args []string, flags *imageFlags) error {
	prompt, err := getPrompt(args, flags.promptFile, cmd.InOrStdin())
	if err != nil {
		return common.WriteError(cmd, "missing_prompt", err.Error())
	}

	if flags.responseFormat != "base64" && flags.responseFormat != "url" {
		return common.WriteError(cmd, "invalid_response_format", "response-format must be base64 or url")
	}

	// -o is required for base64 mode, optional for url mode
	if flags.output == "" && flags.responseFormat == "base64" {
		return common.WriteError(cmd, "missing_output", "output file is required for base64 mode, use -o flag")
	}

	if flags.output != "" {
		ext := strings.ToLower(filepath.Ext(flags.output))
		if ext == "" || !validImageFormats[ext] {
			return common.WriteError(cmd, "unsupported_format", "output file must be png, jpg, jpeg, or webp")
		}
	}

	if flags.count < 1 || flags.count > 9 {
		return common.WriteError(cmd, "invalid_count", "count must be between 1 and 9")
	}

	if flags.width != 0 || flags.height != 0 {
		if flags.width == 0 || flags.height == 0 {
			return common.WriteError(cmd, "invalid_size", "width and height must be set together")
		}
		if flags.width < 512 || flags.width > 2048 || flags.height < 512 || flags.height > 2048 {
			return common.WriteError(cmd, "invalid_size", "width and height must be in range 512-2048")
		}
		if flags.width%8 != 0 || flags.height%8 != 0 {
			return common.WriteError(cmd, "invalid_size", "width and height must be divisible by 8")
		}
	}

	if flags.aspect != "" && !validImageAspects[flags.aspect] {
		return common.WriteError(cmd, "invalid_aspect", fmt.Sprintf("invalid aspect ratio '%s'", flags.aspect))
	}

	if !validImageModels[flags.model] {
		return common.WriteError(cmd, "invalid_model", fmt.Sprintf("invalid model '%s'", flags.model))
	}

	// image-01-live requires reference images
	if flags.model == "image-01-live" && len(flags.images) == 0 {
		return common.WriteError(cmd, "missing_image", "image-01-live requires at least one reference image (-i)")
	}

	apiKey := shared.GetMinimaxAPIKey()
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("MINIMAX_API_KEY"))
	}

	body := map[string]any{
		"model":            flags.model,
		"prompt":           prompt,
		"response_format":  flags.responseFormat,
		"n":                flags.count,
		"prompt_optimizer": flags.promptOptimizer,
	}

	if flags.aspect != "" {
		body["aspect_ratio"] = flags.aspect
	}
	if flags.width != 0 && flags.height != 0 {
		body["width"] = flags.width
		body["height"] = flags.height
	}

	if len(flags.images) > 0 {
		var refs []map[string]any
		for _, img := range flags.images {
			ref, err := shared.ResolveImageURL(img)
			if err != nil {
				return common.WriteError(cmd, "image_read_error", fmt.Sprintf("cannot read image: %s", err.Error()))
			}
			refs = append(refs, map[string]any{
				"type":       "character",
				"image_file": ref,
			})
		}
		body["subject_reference"] = refs
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot serialize request: %s", err.Error()))
	}

	req, err := shared.CreateRequest("POST", "/v1/image_generation", bytes.NewReader(jsonBody))
	if err != nil {
		return common.WriteError(cmd, "request_error", err.Error())
	}

	resp, err := shared.DoRequest(req)
	if err != nil {
		return common.WriteError(cmd, "request_error", err.Error())
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot read response: %s", err.Error()))
	}

	if resp.StatusCode != http.StatusOK {
		return common.WriteError(cmd, "api_error", fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(respBody)))
	}

	var apiResp struct {
		Data struct {
			ImageURLs   []string `json:"image_urls"`
			ImageBase64 []string `json:"image_base64"`
		} `json:"data"`
		BaseResp struct {
			StatusCode int    `json:"status_code"`
			StatusMsg  string `json:"status_msg"`
		} `json:"base_resp"`
	}
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot parse response: %s", err.Error()))
	}

	if apiResp.BaseResp.StatusCode != 0 {
		return common.WriteError(cmd, "api_error", fmt.Sprintf("api error %d: %s", apiResp.BaseResp.StatusCode, apiResp.BaseResp.StatusMsg))
	}

	var results []string
	if flags.responseFormat == "url" {
		results = apiResp.Data.ImageURLs
	} else {
		results = apiResp.Data.ImageBase64
	}

	if len(results) == 0 {
		return common.WriteError(cmd, "no_image", "no image generated in response")
	}

	output := imageResponse{
		Success: true,
		Model:   flags.model,
		Count:   len(results),
	}

	// If url mode without -o, return URLs directly without downloading
	if flags.responseFormat == "url" && flags.output == "" {
		if len(results) == 1 {
			output.URL = results[0]
		} else {
			output.URLs = results
		}
		return common.WriteSuccess(cmd, output)
	}

	// Download and save images
	savedFiles, err := saveImages(flags.output, results, flags.responseFormat == "url")
	if err != nil {
		return common.WriteError(cmd, "output_write_error", err.Error())
	}

	if len(savedFiles) == 1 {
		output.File = savedFiles[0]
	} else {
		output.Files = savedFiles
	}
	return common.WriteSuccess(cmd, output)
}

func saveImages(output string, results []string, isURL bool) ([]string, error) {
	var saved []string

	absPath, err := filepath.Abs(output)
	if err != nil {
		absPath = output
	}

	if len(results) == 1 {
		path := absPath
		if err := writeImage(path, results[0], isURL); err != nil {
			return nil, err
		}
		return []string{path}, nil
	}

	baseName := strings.TrimSuffix(absPath, filepath.Ext(absPath))
	extName := filepath.Ext(absPath)
	for i, item := range results {
		path := fmt.Sprintf("%s_%d%s", baseName, i+1, extName)
		if err := writeImage(path, item, isURL); err != nil {
			return nil, err
		}
		saved = append(saved, path)
	}
	return saved, nil
}

func writeImage(path, data string, isURL bool) error {
	var content []byte
	if isURL {
		client := &http.Client{Timeout: 2 * time.Minute}
		resp, err := client.Get(data)
		if err != nil {
			return fmt.Errorf("cannot download image: %s", err.Error())
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("download failed with status %d", resp.StatusCode)
		}
		content, err = io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("cannot read image: %s", err.Error())
		}
	} else {
		decoded, err := base64.StdEncoding.DecodeString(data)
		if err != nil {
			return fmt.Errorf("cannot decode image: %s", err.Error())
		}
		content = decoded
	}

	if err := os.WriteFile(path, content, 0644); err != nil {
		return fmt.Errorf("cannot write output file: %s", err.Error())
	}
	return nil
}

func getPrompt(args []string, filePath string, stdin io.Reader) (string, error) {
	if len(args) > 0 {
		text := strings.TrimSpace(strings.Join(args, " "))
		if text != "" {
			return text, nil
		}
	}

	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("cannot read file: %s", err.Error())
		}
		text := strings.TrimSpace(string(data))
		if text != "" {
			return text, nil
		}
	}

	if stdin != nil {
		if f, ok := stdin.(*os.File); ok {
			stat, _ := f.Stat()
			if (stat.Mode() & os.ModeCharDevice) != 0 {
				return "", errors.New("no prompt provided")
			}
		}
		data, err := io.ReadAll(stdin)
		if err != nil {
			return "", fmt.Errorf("cannot read stdin: %s", err.Error())
		}
		text := strings.TrimSpace(string(data))
		if text != "" {
			return text, nil
		}
	}

	return "", errors.New("no prompt provided")
}
