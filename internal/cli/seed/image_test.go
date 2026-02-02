package seed

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func executeImageCommand(cmd *cobra.Command, args ...string) (stdout, stderr string, err error) {
	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)

	cmd.SetOut(stdoutBuf)
	cmd.SetErr(stderrBuf)
	cmd.SetArgs(args)

	err = cmd.Execute()
	return stdoutBuf.String(), stderrBuf.String(), err
}

func TestSeedImage_MissingPrompt(t *testing.T) {
	cmd := newSeedImageCmd()
	_, stderr, err := executeImageCommand(cmd, "-o", "output.jpg")

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

func TestSeedImage_MissingOutput(t *testing.T) {
	cmd := newSeedImageCmd()
	_, stderr, err := executeImageCommand(cmd, "A cute cat")

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

func TestSeedImage_UnsupportedFormat(t *testing.T) {
	cmd := newSeedImageCmd()
	_, stderr, err := executeImageCommand(cmd, "A cute cat", "-o", "out.png")

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

func TestSeedImage_InvalidModel(t *testing.T) {
	cmd := newSeedImageCmd()
	_, stderr, err := executeImageCommand(cmd, "A cute cat", "-o", "out.jpg", "--model", "invalid")

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

func TestSeedImage_InvalidSize(t *testing.T) {
	cmd := newSeedImageCmd()
	_, stderr, err := executeImageCommand(cmd, "A cute cat", "-o", "out.jpg", "--size", "3K")

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

func TestSeedImage_InvalidCount(t *testing.T) {
	tests := []struct {
		name  string
		count string
	}{
		{"zero", "0"},
		{"negative", "-1"},
		{"too_large", "11"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newSeedImageCmd()
			_, stderr, err := executeImageCommand(cmd, "A cute cat", "-o", "out.jpg", "-n", tt.count)

			if err == nil {
				t.Fatal("expected error for invalid count")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "invalid_count" {
				t.Errorf("expected error code 'invalid_count', got: %s", errorObj["code"])
			}
		})
	}
}

func TestSeedImage_ImageNotFound(t *testing.T) {
	cmd := newSeedImageCmd()
	_, stderr, err := executeImageCommand(cmd, "Make it better", "--image", "/nonexistent/image.png", "-o", "output.jpg")

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

func TestSeedImage_TooManyImages(t *testing.T) {
	// Create 15 temp files (max is 14)
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

	args := []string{"Combine all", "-o", "output.jpg"}
	for _, f := range tmpFiles {
		args = append(args, "--image", f.Name())
	}

	cmd := newSeedImageCmd()
	_, stderr, err := executeImageCommand(cmd, args...)

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

func TestSeedImage_MissingAPIKey(t *testing.T) {
	t.Setenv("ARK_API_KEY", "")

	cmd := newSeedImageCmd()
	_, stderr, err := executeImageCommand(cmd, "A cute cat", "-o", "out.jpg")

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

func TestSeedImage_ValidFlags(t *testing.T) {
	cmd := newSeedImageCmd()

	flags := []string{"output", "image", "prompt-file", "model", "size", "count", "watermark"}
	for _, flag := range flags {
		if cmd.Flag(flag) == nil {
			t.Errorf("expected --%s flag", flag)
		}
	}
}

func TestSeedImage_DefaultValues(t *testing.T) {
	cmd := newSeedImageCmd()

	if cmd.Flag("model").DefValue != "4.5" {
		t.Errorf("expected default model '4.5', got: %s", cmd.Flag("model").DefValue)
	}
	if cmd.Flag("size").DefValue != "2K" {
		t.Errorf("expected default size '2K', got: %s", cmd.Flag("size").DefValue)
	}
	if cmd.Flag("count").DefValue != "1" {
		t.Errorf("expected default count '1', got: %s", cmd.Flag("count").DefValue)
	}
	if cmd.Flag("watermark").DefValue != "false" {
		t.Errorf("expected default watermark 'false', got: %s", cmd.Flag("watermark").DefValue)
	}
}

func TestSeedImage_ShortFlags(t *testing.T) {
	cmd := newSeedImageCmd()

	shortFlags := map[string]string{
		"o": "output",
		"i": "image",
		"m": "model",
		"s": "size",
		"n": "count",
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

func TestSeedImage_ValidSizes(t *testing.T) {
	t.Setenv("ARK_API_KEY", "")

	sizes := []string{"2K", "4K"}
	for _, size := range sizes {
		t.Run(size, func(t *testing.T) {
			cmd := newSeedImageCmd()
			_, stderr, err := executeImageCommand(cmd, "A cute cat", "-o", "out.jpg", "--size", size)

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

func TestSeedImage_ValidWxHSizes(t *testing.T) {
	t.Setenv("ARK_API_KEY", "")

	sizes := []string{"2048x2048", "1024x1024", "3072x2048", "1920x1080"}
	for _, size := range sizes {
		t.Run(size, func(t *testing.T) {
			cmd := newSeedImageCmd()
			_, stderr, err := executeImageCommand(cmd, "A cute cat", "-o", "out.jpg", "--size", size)

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

func TestSeedImage_InvalidWxHSizes(t *testing.T) {
	sizes := []string{
		"10x10",     // too small pixels
		"100x100",   // too small total pixels
		"5000x5000", // too large total pixels
	}
	for _, size := range sizes {
		t.Run(size, func(t *testing.T) {
			cmd := newSeedImageCmd()
			_, stderr, err := executeImageCommand(cmd, "A cute cat", "-o", "out.jpg", "--size", size)

			if err == nil {
				t.Fatal("expected error for invalid size")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "invalid_size" {
				t.Errorf("expected error code 'invalid_size' for size '%s', got: %s", size, errorObj["code"])
			}
		})
	}
}

func TestSeedImage_ValidModels(t *testing.T) {
	t.Setenv("ARK_API_KEY", "")

	models := []string{"4.5", "4.0"}
	for _, model := range models {
		t.Run(model, func(t *testing.T) {
			cmd := newSeedImageCmd()
			_, stderr, err := executeImageCommand(cmd, "A cute cat", "-o", "out.jpg", "--model", model)

			if err == nil {
				t.Fatal("expected error (missing api key), got success")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "missing_api_key" {
				t.Errorf("expected model '%s' to be valid, got error: %s", model, errorObj["code"])
			}
		})
	}
}

func TestSeedImage_FromFile(t *testing.T) {
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

	t.Setenv("ARK_API_KEY", "")

	cmd := newSeedImageCmd()
	_, stderr, err := executeImageCommand(cmd, "--prompt-file", tmpFile.Name(), "-o", "out.jpg")

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

func TestSeedImage_FromStdin(t *testing.T) {
	t.Setenv("ARK_API_KEY", "")

	cmd := newSeedImageCmd()
	cmd.SetIn(strings.NewReader("A beautiful landscape"))

	_, stderr, err := executeImageCommand(cmd, "-o", "out.jpg")

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

func TestSeedImage_WithReferenceImage(t *testing.T) {
	t.Setenv("ARK_API_KEY", "")

	tmpFile, err := os.CreateTemp("", "image_*.png")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cmd := newSeedImageCmd()
	_, stderr, err := executeImageCommand(cmd, "Make it better", "--image", tmpFile.Name(), "-o", "out.jpg")

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

func TestSeedImage_PromptFileNotFound(t *testing.T) {
	cmd := newSeedImageCmd()
	_, stderr, err := executeImageCommand(cmd, "--prompt-file", "/nonexistent/prompt.txt", "-o", "out.jpg")

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

func TestSeedImage_EmptyPromptFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "prompt_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	_, _ = tmpFile.WriteString("   \n\t  ")
	tmpFile.Close()

	cmd := newSeedImageCmd()
	_, stderr, err := executeImageCommand(cmd, "--prompt-file", tmpFile.Name(), "-o", "out.jpg")

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

func TestSeedImage_EmptyStringPrompt(t *testing.T) {
	cmd := newSeedImageCmd()
	_, stderr, err := executeImageCommand(cmd, "", "-o", "out.jpg")

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

func TestSeedImage_WhitespaceOnlyPrompt(t *testing.T) {
	cmd := newSeedImageCmd()
	_, stderr, err := executeImageCommand(cmd, "   \t\n  ", "-o", "out.jpg")

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

func TestSeedImage_EmptyStdin(t *testing.T) {
	cmd := newSeedImageCmd()
	cmd.SetIn(strings.NewReader(""))

	_, stderr, err := executeImageCommand(cmd, "-o", "out.jpg")

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

func TestSeedImage_MaxImagesAllowed(t *testing.T) {
	t.Setenv("ARK_API_KEY", "")

	// Create 14 temp files (max is 14, should pass)
	var tmpFiles []*os.File
	for i := 0; i < 14; i++ {
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

	args := []string{"Combine all", "-o", "output.jpg"}
	for _, f := range tmpFiles {
		args = append(args, "--image", f.Name())
	}

	cmd := newSeedImageCmd()
	_, stderr, err := executeImageCommand(cmd, args...)

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
		t.Errorf("expected 14 images to be valid, got error: %s", errorObj["code"])
	}
}

func TestSeedImage_OutputUppercaseJPG(t *testing.T) {
	t.Setenv("ARK_API_KEY", "")

	cmd := newSeedImageCmd()
	_, stderr, err := executeImageCommand(cmd, "A cute cat", "-o", "out.JPG")

	if err == nil {
		t.Fatal("expected error (missing api key)")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	// Should accept uppercase .JPG and reach API key check
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected .JPG to be valid, got error: %s", errorObj["code"])
	}
}

func TestSeedImage_OutputJPEG(t *testing.T) {
	t.Setenv("ARK_API_KEY", "")

	cmd := newSeedImageCmd()
	_, stderr, err := executeImageCommand(cmd, "A cute cat", "-o", "out.jpeg")

	if err == nil {
		t.Fatal("expected error (missing api key)")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	// Should accept .jpeg and reach API key check
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected .jpeg to be valid, got error: %s", errorObj["code"])
	}
}

func TestSeedImage_ValidCounts(t *testing.T) {
	t.Setenv("ARK_API_KEY", "")

	counts := []string{"1", "5", "10"}
	for _, count := range counts {
		t.Run(count, func(t *testing.T) {
			cmd := newSeedImageCmd()
			_, stderr, err := executeImageCommand(cmd, "A cute cat", "-o", "out.jpg", "-n", count)

			if err == nil {
				t.Fatal("expected error (missing api key), got success")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "missing_api_key" {
				t.Errorf("expected count '%s' to be valid, got error: %s", count, errorObj["code"])
			}
		})
	}
}

func TestSeedImage_WithWatermark(t *testing.T) {
	t.Setenv("ARK_API_KEY", "")

	cmd := newSeedImageCmd()
	_, stderr, err := executeImageCommand(cmd, "A cute cat", "-o", "out.jpg", "--watermark")

	if err == nil {
		t.Fatal("expected error (missing api key), got success")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected watermark flag to be valid, got error: %s", errorObj["code"])
	}
}

func TestGetSeedMimeType(t *testing.T) {
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
		{"image.bmp", "image/bmp"},
		{"image.BMP", "image/bmp"},
		{"image.tiff", "image/tiff"},
		{"image.tif", "image/tiff"},
		{"image.unknown", "application/octet-stream"},
		{"image", "application/octet-stream"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := getSeedMimeType(tt.path)
			if result != tt.expected {
				t.Errorf("getSeedMimeType(%q) = %q, want %q", tt.path, result, tt.expected)
			}
		})
	}
}

func TestIsValidSeedSize(t *testing.T) {
	tests := []struct {
		size     string
		expected bool
	}{
		{"2K", true},
		{"4K", true},
		{"1K", false},
		{"3K", false},
		{"2048x2048", true},
		{"1024x1024", true},
		{"3072x2048", true},
		{"1920x1080", true},
		{"10x10", false},       // too small
		{"100x100", false},     // too small total pixels
		{"5000x5000", false},   // too large
		{"invalid", false},
		{"2048", false},
		{"x2048", false},
		{"2048x", false},
	}

	for _, tt := range tests {
		t.Run(tt.size, func(t *testing.T) {
			result := isValidSeedSize(tt.size)
			if result != tt.expected {
				t.Errorf("isValidSeedSize(%q) = %v, want %v", tt.size, result, tt.expected)
			}
		})
	}
}
