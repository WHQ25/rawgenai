package openai

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

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

func TestImage_InvalidCompression(t *testing.T) {
	tests := []struct {
		name        string
		compression string
	}{
		{"negative", "-1"},
		{"too large", "101"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newImageCmd()
			_, stderr, err := executeCommand(cmd, "A cute cat", "-o", "out.jpeg", "--compression", tt.compression)

			if err == nil {
				t.Fatal("expected error for invalid compression")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "invalid_compression" {
				t.Errorf("expected error code 'invalid_compression', got: %s", errorObj["code"])
			}
		})
	}
}

func TestImage_InvalidFidelity(t *testing.T) {
	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "A cute cat", "-o", "out.png", "--fidelity", "medium")

	if err == nil {
		t.Fatal("expected error for invalid fidelity")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_fidelity" {
		t.Errorf("expected error code 'invalid_fidelity', got: %s", errorObj["code"])
	}
}

func TestImage_TransparentRequiresPngWebp(t *testing.T) {
	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "A cute cat", "-o", "out.jpeg", "--background", "transparent")

	if err == nil {
		t.Fatal("expected error for transparent with jpeg")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "transparent_requires_png_webp" {
		t.Errorf("expected error code 'transparent_requires_png_webp', got: %s", errorObj["code"])
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
	if errorObj["code"] != "file_not_found" {
		t.Errorf("expected error code 'file_not_found', got: %s", errorObj["code"])
	}
}

func TestImage_MaskNotFound(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "image_*.png")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "Edit this", "--image", tmpFile.Name(), "--mask", "/nonexistent/mask.png", "-o", "output.png")

	if err == nil {
		t.Fatal("expected error for mask not found")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "file_not_found" {
		t.Errorf("expected error code 'file_not_found', got: %s", errorObj["code"])
	}
}

func TestImage_MaskRequiresImage(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "mask_*.png")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "Edit this", "--mask", tmpFile.Name(), "-o", "output.png")

	if err == nil {
		t.Fatal("expected error for mask without image")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "mask_requires_image" {
		t.Errorf("expected error code 'mask_requires_image', got: %s", errorObj["code"])
	}
}

func TestImage_TooManyImages(t *testing.T) {
	// Create 17 temp files
	var tmpFiles []*os.File
	for i := 0; i < 17; i++ {
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

	args := []string{"Combine all", "-o", "output.png"}
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
	t.Setenv("OPENAI_API_KEY", "")

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

	flags := []string{"output", "image", "mask", "file", "continue", "model", "size", "quality", "background", "compression", "fidelity", "moderation"}
	for _, flag := range flags {
		if cmd.Flag(flag) == nil {
			t.Errorf("expected --%s flag", flag)
		}
	}
}

func TestImage_ContinueWithResponseID(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")

	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "Add more details", "--continue", "resp_abc123", "-o", "out.png")

	if err == nil {
		t.Fatal("expected error (missing api key), got success")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	// Should reach API key check, meaning --continue flag was accepted
	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key' (continue flag accepted), got: %s", errorObj["code"])
	}
}

func TestImage_DefaultValues(t *testing.T) {
	cmd := newImageCmd()

	if cmd.Flag("model").DefValue != "gpt-image-1" {
		t.Errorf("expected default model 'gpt-image-1', got: %s", cmd.Flag("model").DefValue)
	}
	if cmd.Flag("size").DefValue != "auto" {
		t.Errorf("expected default size 'auto', got: %s", cmd.Flag("size").DefValue)
	}
	if cmd.Flag("quality").DefValue != "auto" {
		t.Errorf("expected default quality 'auto', got: %s", cmd.Flag("quality").DefValue)
	}
	if cmd.Flag("background").DefValue != "auto" {
		t.Errorf("expected default background 'auto', got: %s", cmd.Flag("background").DefValue)
	}
	if cmd.Flag("compression").DefValue != "100" {
		t.Errorf("expected default compression '100', got: %s", cmd.Flag("compression").DefValue)
	}
	if cmd.Flag("fidelity").DefValue != "low" {
		t.Errorf("expected default fidelity 'low', got: %s", cmd.Flag("fidelity").DefValue)
	}
	if cmd.Flag("moderation").DefValue != "auto" {
		t.Errorf("expected default moderation 'auto', got: %s", cmd.Flag("moderation").DefValue)
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

	t.Setenv("OPENAI_API_KEY", "")

	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "--file", tmpFile.Name(), "-o", "out.png")

	if err == nil {
		t.Fatal("expected error (missing api key), got success")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	// Should reach API key check, meaning file was read successfully
	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key' (file read success), got: %s", errorObj["code"])
	}
}

func TestImage_FromStdin(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")

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
	t.Setenv("OPENAI_API_KEY", "")

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

	// Should reach API key check, meaning image file was validated successfully
	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key' (image validated), got: %s", errorObj["code"])
	}
}

func TestImage_MultipleReferenceImages(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")

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

	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "Combine these", "--image", tmpFiles[0].Name(), "--image", tmpFiles[1].Name(), "--image", tmpFiles[2].Name(), "-o", "out.png")

	if err == nil {
		t.Fatal("expected error (missing api key), got success")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key' (multiple images validated), got: %s", errorObj["code"])
	}
}

func TestImage_SupportedFormats(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")

	formats := []string{".png", ".jpeg", ".jpg", ".webp"}
	for _, ext := range formats {
		t.Run(ext, func(t *testing.T) {
			cmd := newImageCmd()
			_, stderr, err := executeCommand(cmd, "A cute cat", "-o", "out"+ext)

			if err == nil {
				t.Fatal("expected error (missing api key), got success")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			// Should pass format validation and reach API key check
			if errorObj["code"] != "missing_api_key" {
				t.Errorf("expected format '%s' to be supported, got error: %s", ext, errorObj["code"])
			}
		})
	}
}

func TestImage_ShortFlags(t *testing.T) {
	cmd := newImageCmd()

	shortFlags := map[string]string{
		"o": "output",
		"i": "image",
		"c": "continue",
		"m": "model",
		"s": "size",
		"q": "quality",
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
	// Write empty content (just whitespace)
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

func TestImage_TransparentWithWebp(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")

	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "A cute cat", "-o", "out.webp", "--background", "transparent")

	if err == nil {
		t.Fatal("expected error (missing api key), got success")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	// Should pass validation (webp supports transparency) and reach API key check
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected transparent+webp to be valid, got error: %s", errorObj["code"])
	}
}

func TestImage_CompressionBoundary(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")

	tests := []struct {
		name        string
		compression string
	}{
		{"zero", "0"},
		{"hundred", "100"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newImageCmd()
			_, stderr, err := executeCommand(cmd, "A cute cat", "-o", "out.jpeg", "--compression", tt.compression)

			if err == nil {
				t.Fatal("expected error (missing api key), got success")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "missing_api_key" {
				t.Errorf("expected compression %s to be valid, got error: %s", tt.compression, errorObj["code"])
			}
		})
	}
}

func TestImage_ValidFidelityValues(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")

	values := []string{"high", "low"}
	for _, fidelity := range values {
		t.Run(fidelity, func(t *testing.T) {
			cmd := newImageCmd()
			_, stderr, err := executeCommand(cmd, "A cute cat", "-o", "out.png", "--fidelity", fidelity)

			if err == nil {
				t.Fatal("expected error (missing api key), got success")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "missing_api_key" {
				t.Errorf("expected fidelity '%s' to be valid, got error: %s", fidelity, errorObj["code"])
			}
		})
	}
}

func TestImage_MaskWithImage(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")

	imgFile, err := os.CreateTemp("", "image_*.png")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(imgFile.Name())
	imgFile.Close()

	maskFile, err := os.CreateTemp("", "mask_*.png")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(maskFile.Name())
	maskFile.Close()

	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "Add a flamingo", "--image", imgFile.Name(), "--mask", maskFile.Name(), "-o", "out.png")

	if err == nil {
		t.Fatal("expected error (missing api key), got success")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	// Should pass validation and reach API key check
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected mask+image to be valid, got error: %s", errorObj["code"])
	}
}
