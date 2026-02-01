package google

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func executeCommand(cmd *cobra.Command, args ...string) (stdout, stderr string, err error) {
	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)

	cmd.SetOut(stdoutBuf)
	cmd.SetErr(stderrBuf)
	cmd.SetArgs(args)

	err = cmd.Execute()
	return stdoutBuf.String(), stderrBuf.String(), err
}

func TestImage_MissingPrompt(t *testing.T) {
	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "-o", "output.png")

	if err == nil {
		t.Fatal("expected error for missing prompt")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	if resp["success"] != false {
		t.Error("expected success to be false")
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_prompt" {
		t.Errorf("expected error code 'missing_prompt', got: %s", errorObj["code"])
	}
}

func TestImage_MissingOutput(t *testing.T) {
	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "A cute cat")

	if err == nil {
		t.Fatal("expected error for missing output")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_output" {
		t.Errorf("expected error code 'missing_output', got: %s", errorObj["code"])
	}
}

func TestImage_UnsupportedFormat(t *testing.T) {
	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "A cute cat", "-o", "out.gif")

	if err == nil {
		t.Fatal("expected error for unsupported format")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "unsupported_format" {
		t.Errorf("expected error code 'unsupported_format', got: %s", errorObj["code"])
	}
}

func TestImage_InvalidModel(t *testing.T) {
	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "A cute cat", "-o", "out.png", "--model", "invalid")

	if err == nil {
		t.Fatal("expected error for invalid model")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_model" {
		t.Errorf("expected error code 'invalid_model', got: %s", errorObj["code"])
	}
}

func TestImage_InvalidAspect(t *testing.T) {
	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "A cute cat", "-o", "out.png", "--aspect", "1:2")

	if err == nil {
		t.Fatal("expected error for invalid aspect ratio")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_aspect" {
		t.Errorf("expected error code 'invalid_aspect', got: %s", errorObj["code"])
	}
}

func TestImage_InvalidSize(t *testing.T) {
	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "A cute cat", "-o", "out.png", "--model", "pro", "--size", "3K")

	if err == nil {
		t.Fatal("expected error for invalid size")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_size" {
		t.Errorf("expected error code 'invalid_size', got: %s", errorObj["code"])
	}
}

func TestImage_SizeRequiresPro(t *testing.T) {
	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "A cute cat", "-o", "out.png", "--model", "flash", "--size", "2K")

	if err == nil {
		t.Fatal("expected error for size with flash model")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "size_requires_pro" {
		t.Errorf("expected error code 'size_requires_pro', got: %s", errorObj["code"])
	}
}

func TestImage_SearchRequiresPro(t *testing.T) {
	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "Current weather", "-o", "out.png", "--model", "flash", "--search")

	if err == nil {
		t.Fatal("expected error for search with flash model")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "search_requires_pro" {
		t.Errorf("expected error code 'search_requires_pro', got: %s", errorObj["code"])
	}
}

func TestImage_ImageNotFound(t *testing.T) {
	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "Make it better", "--image", "/nonexistent/image.png", "-o", "output.png")

	if err == nil {
		t.Fatal("expected error for image not found")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "image_not_found" {
		t.Errorf("expected error code 'image_not_found', got: %s", errorObj["code"])
	}
}

func TestImage_TooManyImagesFlash(t *testing.T) {
	// Create 4 temp files (flash max is 3)
	var tmpFiles []*os.File
	for i := 0; i < 4; i++ {
		tmpFile, err := os.CreateTemp("", "image_*.png")
		if err != nil {
			t.Fatal(err)
		}
		tmpFiles = append(tmpFiles, tmpFile)
		tmpFile.Close()
	}
	defer func() {
		for _, f := range tmpFiles {
			os.Remove(f.Name())
		}
	}()

	args := []string{"Combine all", "-o", "output.png", "--model", "flash"}
	for _, f := range tmpFiles {
		args = append(args, "--image", f.Name())
	}

	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, args...)

	if err == nil {
		t.Fatal("expected error for too many images")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "too_many_images" {
		t.Errorf("expected error code 'too_many_images', got: %s", errorObj["code"])
	}
}

func TestImage_TooManyImagesPro(t *testing.T) {
	// Create 15 temp files (pro max is 14)
	var tmpFiles []*os.File
	for i := 0; i < 15; i++ {
		tmpFile, err := os.CreateTemp("", "image_*.png")
		if err != nil {
			t.Fatal(err)
		}
		tmpFiles = append(tmpFiles, tmpFile)
		tmpFile.Close()
	}
	defer func() {
		for _, f := range tmpFiles {
			os.Remove(f.Name())
		}
	}()

	args := []string{"Combine all", "-o", "output.png", "--model", "pro"}
	for _, f := range tmpFiles {
		args = append(args, "--image", f.Name())
	}

	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, args...)

	if err == nil {
		t.Fatal("expected error for too many images")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "too_many_images" {
		t.Errorf("expected error code 'too_many_images', got: %s", errorObj["code"])
	}
}

func TestImage_MissingAPIKey(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GOOGLE_API_KEY", "")

	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "A cute cat", "-o", "out.png")

	if err == nil {
		t.Fatal("expected error for missing API key")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key', got: %s", errorObj["code"])
	}
}

func TestImage_ValidFlags(t *testing.T) {
	cmd := newImageCmd()

	flags := []string{"output", "image", "file", "model", "aspect", "size", "search"}
	for _, flag := range flags {
		if cmd.Flag(flag) == nil {
			t.Errorf("expected --%s flag", flag)
		}
	}
}

func TestImage_DefaultValues(t *testing.T) {
	cmd := newImageCmd()

	if cmd.Flag("model").DefValue != "flash" {
		t.Errorf("expected default model 'flash', got: %s", cmd.Flag("model").DefValue)
	}
	if cmd.Flag("aspect").DefValue != "1:1" {
		t.Errorf("expected default aspect '1:1', got: %s", cmd.Flag("aspect").DefValue)
	}
	if cmd.Flag("size").DefValue != "1K" {
		t.Errorf("expected default size '1K', got: %s", cmd.Flag("size").DefValue)
	}
	if cmd.Flag("search").DefValue != "false" {
		t.Errorf("expected default search 'false', got: %s", cmd.Flag("search").DefValue)
	}
}

func TestImage_ShortFlags(t *testing.T) {
	cmd := newImageCmd()

	shortFlags := map[string]string{
		"o": "output",
		"i": "image",
		"m": "model",
		"a": "aspect",
		"s": "size",
	}

	for short, long := range shortFlags {
		flag := cmd.Flag(long)
		if flag == nil {
			t.Errorf("flag --%s not found", long)
			continue
		}
		if flag.Shorthand != short {
			t.Errorf("expected short flag '-%s' for '--%s', got '-%s'", short, long, flag.Shorthand)
		}
	}
}

func TestImage_ValidAspectRatios(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")

	aspects := []string{"1:1", "2:3", "3:2", "3:4", "4:3", "4:5", "5:4", "9:16", "16:9", "21:9"}
	for _, aspect := range aspects {
		t.Run(aspect, func(t *testing.T) {
			cmd := newImageCmd()
			_, stderr, err := executeCommand(cmd, "A cute cat", "-o", "out.png", "--aspect", aspect)

			if err == nil {
				t.Fatal("expected error (missing api key), got success")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			// Should pass aspect validation and reach API key check
			if errorObj["code"] != "missing_api_key" {
				t.Errorf("expected aspect '%s' to be valid, got error: %s", aspect, errorObj["code"])
			}
		})
	}
}

func TestImage_ValidSizes(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")

	sizes := []string{"1K", "2K", "4K"}
	for _, size := range sizes {
		t.Run(size, func(t *testing.T) {
			cmd := newImageCmd()
			_, stderr, err := executeCommand(cmd, "A cute cat", "-o", "out.png", "--model", "pro", "--size", size)

			if err == nil {
				t.Fatal("expected error (missing api key), got success")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "missing_api_key" {
				t.Errorf("expected size '%s' to be valid, got error: %s", size, errorObj["code"])
			}
		})
	}
}

func TestImage_LowercaseSizeRejected(t *testing.T) {
	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "A cute cat", "-o", "out.png", "--model", "pro", "--size", "2k")

	if err == nil {
		t.Fatal("expected error for lowercase size")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_size" {
		t.Errorf("expected error code 'invalid_size', got: %s", errorObj["code"])
	}
}

func TestImage_FromFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "prompt_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString("A beautiful landscape")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	t.Setenv("GEMINI_API_KEY", "")

	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "--file", tmpFile.Name(), "-o", "out.png")

	if err == nil {
		t.Fatal("expected error (missing api key), got success")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key' (file read success), got: %s", errorObj["code"])
	}
}

func TestImage_FromStdin(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")

	cmd := newImageCmd()
	cmd.SetIn(strings.NewReader("A beautiful landscape"))

	_, stderr, err := executeCommand(cmd, "-o", "out.png")

	if err == nil {
		t.Fatal("expected error (missing api key), got success")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key' (stdin read success), got: %s", errorObj["code"])
	}
}

func TestImage_WithReferenceImage(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")

	tmpFile, err := os.CreateTemp("", "image_*.png")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "Make it better", "--image", tmpFile.Name(), "-o", "out.png")

	if err == nil {
		t.Fatal("expected error (missing api key), got success")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key' (image validated), got: %s", errorObj["code"])
	}
}

func TestImage_ProModelWithSearch(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")

	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "Weather in Tokyo", "-o", "out.png", "--model", "pro", "--search")

	if err == nil {
		t.Fatal("expected error (missing api key), got success")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected pro+search to be valid, got error: %s", errorObj["code"])
	}
}

func TestImage_PromptFileNotFound(t *testing.T) {
	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "--file", "/nonexistent/prompt.txt", "-o", "out.png")

	if err == nil {
		t.Fatal("expected error for prompt file not found")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_prompt" {
		t.Errorf("expected error code 'missing_prompt', got: %s", errorObj["code"])
	}
}

func TestImage_EmptyPromptFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "prompt_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	_, _ = tmpFile.WriteString("   \n\t  ")
	tmpFile.Close()

	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "--file", tmpFile.Name(), "-o", "out.png")

	if err == nil {
		t.Fatal("expected error for empty prompt file")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_prompt" {
		t.Errorf("expected error code 'missing_prompt', got: %s", errorObj["code"])
	}
}

func TestImage_EmptyStringPrompt(t *testing.T) {
	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "", "-o", "out.png")

	if err == nil {
		t.Fatal("expected error for empty string prompt")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_prompt" {
		t.Errorf("expected error code 'missing_prompt', got: %s", errorObj["code"])
	}
}

func TestImage_WhitespaceOnlyPrompt(t *testing.T) {
	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "   \t\n  ", "-o", "out.png")

	if err == nil {
		t.Fatal("expected error for whitespace-only prompt")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_prompt" {
		t.Errorf("expected error code 'missing_prompt', got: %s", errorObj["code"])
	}
}

func TestImage_EmptyStdin(t *testing.T) {
	cmd := newImageCmd()
	cmd.SetIn(strings.NewReader(""))

	_, stderr, err := executeCommand(cmd, "-o", "out.png")

	if err == nil {
		t.Fatal("expected error for empty stdin")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_prompt" {
		t.Errorf("expected error code 'missing_prompt', got: %s", errorObj["code"])
	}
}

func TestImage_MaxImagesFlashAllowed(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GOOGLE_API_KEY", "")

	// Create 3 temp files (flash max is 3, should pass)
	var tmpFiles []*os.File
	for i := 0; i < 3; i++ {
		tmpFile, err := os.CreateTemp("", "image_*.png")
		if err != nil {
			t.Fatal(err)
		}
		tmpFiles = append(tmpFiles, tmpFile)
		tmpFile.Close()
	}
	defer func() {
		for _, f := range tmpFiles {
			os.Remove(f.Name())
		}
	}()

	args := []string{"Combine all", "-o", "output.png", "--model", "flash"}
	for _, f := range tmpFiles {
		args = append(args, "--image", f.Name())
	}

	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, args...)

	if err == nil {
		t.Fatal("expected error (missing api key)")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	// Should pass image count validation and reach API key check
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected 3 images to be valid for flash, got error: %s", errorObj["code"])
	}
}

func TestImage_OutputUppercasePNG(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GOOGLE_API_KEY", "")

	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "A cute cat", "-o", "out.PNG")

	if err == nil {
		t.Fatal("expected error (missing api key)")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	// Should accept uppercase .PNG and reach API key check
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected .PNG to be valid, got error: %s", errorObj["code"])
	}
}

func TestImage_GoogleAPIKeyFallback(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GOOGLE_API_KEY", "test-google-key")

	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "A cute cat", "-o", "out.png")

	// Should not get missing_api_key error since GOOGLE_API_KEY is set
	// Will fail at API call, but that's expected
	if err == nil {
		return // Unlikely, but if API accepts test key, that's fine
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] == "missing_api_key" {
		t.Error("GOOGLE_API_KEY should be used as fallback")
	}
}

func TestGetMimeType(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"image.png", "image/png"},
		{"image.PNG", "image/png"},
		{"image.jpg", "image/jpeg"},
		{"image.JPG", "image/jpeg"},
		{"image.jpeg", "image/jpeg"},
		{"image.JPEG", "image/jpeg"},
		{"image.webp", "image/webp"},
		{"image.WEBP", "image/webp"},
		{"image.gif", "image/gif"},
		{"image.GIF", "image/gif"},
		{"image.unknown", "application/octet-stream"},
		{"image", "application/octet-stream"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := getMimeType(tt.path)
			if result != tt.expected {
				t.Errorf("getMimeType(%q) = %q, want %q", tt.path, result, tt.expected)
			}
		})
	}
}

func TestHandleAPIError_InvalidAPIKey(t *testing.T) {
	cmd := &cobra.Command{}
	stderrBuf := new(bytes.Buffer)
	cmd.SetErr(stderrBuf)

	err := handleAPIError(cmd, fmt.Errorf("401 Unauthorized: invalid key"))

	if err == nil {
		t.Fatal("expected error")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderrBuf.String())), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderrBuf.String())
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_api_key" {
		t.Errorf("expected 'invalid_api_key', got: %s", errorObj["code"])
	}
}

func TestHandleAPIError_PermissionDenied(t *testing.T) {
	cmd := &cobra.Command{}
	stderrBuf := new(bytes.Buffer)
	cmd.SetErr(stderrBuf)

	err := handleAPIError(cmd, fmt.Errorf("403 Forbidden: permission denied"))

	if err == nil {
		t.Fatal("expected error")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderrBuf.String())), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderrBuf.String())
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "permission_denied" {
		t.Errorf("expected 'permission_denied', got: %s", errorObj["code"])
	}
}

func TestHandleAPIError_QuotaExceeded(t *testing.T) {
	cmd := &cobra.Command{}
	stderrBuf := new(bytes.Buffer)
	cmd.SetErr(stderrBuf)

	err := handleAPIError(cmd, fmt.Errorf("429: quota exceeded"))

	if err == nil {
		t.Fatal("expected error")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderrBuf.String())), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderrBuf.String())
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "quota_exceeded" {
		t.Errorf("expected 'quota_exceeded', got: %s", errorObj["code"])
	}
}

func TestHandleAPIError_RateLimit(t *testing.T) {
	cmd := &cobra.Command{}
	stderrBuf := new(bytes.Buffer)
	cmd.SetErr(stderrBuf)

	err := handleAPIError(cmd, fmt.Errorf("429 Too Many Requests"))

	if err == nil {
		t.Fatal("expected error")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderrBuf.String())), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderrBuf.String())
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "rate_limit" {
		t.Errorf("expected 'rate_limit', got: %s", errorObj["code"])
	}
}

func TestHandleAPIError_ContentPolicy(t *testing.T) {
	cmd := &cobra.Command{}
	stderrBuf := new(bytes.Buffer)
	cmd.SetErr(stderrBuf)

	err := handleAPIError(cmd, fmt.Errorf("content blocked by safety policy"))

	if err == nil {
		t.Fatal("expected error")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderrBuf.String())), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderrBuf.String())
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "content_policy" {
		t.Errorf("expected 'content_policy', got: %s", errorObj["code"])
	}
}

func TestHandleAPIError_Timeout(t *testing.T) {
	cmd := &cobra.Command{}
	stderrBuf := new(bytes.Buffer)
	cmd.SetErr(stderrBuf)

	err := handleAPIError(cmd, fmt.Errorf("request timeout"))

	if err == nil {
		t.Fatal("expected error")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderrBuf.String())), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderrBuf.String())
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "timeout" {
		t.Errorf("expected 'timeout', got: %s", errorObj["code"])
	}
}

func TestHandleAPIError_Connection(t *testing.T) {
	cmd := &cobra.Command{}
	stderrBuf := new(bytes.Buffer)
	cmd.SetErr(stderrBuf)

	err := handleAPIError(cmd, fmt.Errorf("connection refused"))

	if err == nil {
		t.Fatal("expected error")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderrBuf.String())), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderrBuf.String())
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "connection_error" {
		t.Errorf("expected 'connection_error', got: %s", errorObj["code"])
	}
}

func TestHandleAPIError_Generic(t *testing.T) {
	cmd := &cobra.Command{}
	stderrBuf := new(bytes.Buffer)
	cmd.SetErr(stderrBuf)

	err := handleAPIError(cmd, fmt.Errorf("some unknown error"))

	if err == nil {
		t.Fatal("expected error")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderrBuf.String())), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderrBuf.String())
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "api_error" {
		t.Errorf("expected 'api_error', got: %s", errorObj["code"])
	}
}
