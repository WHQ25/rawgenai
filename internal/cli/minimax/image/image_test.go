package image

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestImage_MissingPrompt(t *testing.T) {
	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "-o", "out.png")

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
}

func TestImage_MissingOutputBase64(t *testing.T) {
	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "A cat")

	if err == nil {
		t.Fatal("expected error for missing output in base64 mode")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errObj := resp["error"].(map[string]any)
	if errObj["code"] != "missing_output" {
		t.Errorf("expected error code missing_output, got: %v", errObj["code"])
	}
}

func TestImage_URLModeNoOutputAllowed(t *testing.T) {
	// Temporarily unset API key to prevent API call
	oldKey := os.Getenv("MINIMAX_API_KEY")
	os.Unsetenv("MINIMAX_API_KEY")
	defer func() {
		if oldKey != "" {
			os.Setenv("MINIMAX_API_KEY", oldKey)
		}
	}()

	// url mode without -o should not fail at validation stage
	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "A cat", "--response-format", "url")

	if err == nil {
		t.Skip("API key found in config, cannot test validation bypass")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errObj := resp["error"].(map[string]any)
	// Should NOT fail at missing_output validation for url mode
	if errObj["code"] == "missing_output" {
		t.Error("url mode should not require -o flag")
	}
}

func TestImage_Base64ModeRequiresOutput(t *testing.T) {
	// base64 mode (default) without -o should fail with missing_output
	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "A cat", "--response-format", "base64")

	if err == nil {
		t.Fatal("expected error for missing output in base64 mode")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errObj := resp["error"].(map[string]any)
	if errObj["code"] != "missing_output" {
		t.Errorf("expected error code missing_output, got: %v", errObj["code"])
	}
}

func TestImage_InvalidModel(t *testing.T) {
	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "A cat", "-o", "out.png", "-m", "invalid-model")

	if err == nil {
		t.Fatal("expected error for invalid model")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errObj := resp["error"].(map[string]any)
	if errObj["code"] != "invalid_model" {
		t.Errorf("expected error code invalid_model, got: %v", errObj["code"])
	}
}

func TestImage_LiveModelRequiresImage(t *testing.T) {
	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "A cat", "-o", "out.png", "-m", "image-01-live")

	if err == nil {
		t.Fatal("expected error for missing image with image-01-live")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errObj := resp["error"].(map[string]any)
	if errObj["code"] != "missing_image" {
		t.Errorf("expected error code missing_image, got: %v", errObj["code"])
	}
}

func TestImage_InvalidCount(t *testing.T) {
	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "A cat", "-o", "out.png", "-n", "0")

	if err == nil {
		t.Fatal("expected error for invalid count")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}
	errObj := resp["error"].(map[string]any)
	if errObj["code"] != "invalid_count" {
		t.Errorf("expected error code invalid_count, got: %v", errObj["code"])
	}
}

func TestImage_InvalidCountMax(t *testing.T) {
	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "A cat", "-o", "out.png", "-n", "10")

	if err == nil {
		t.Fatal("expected error for count > 9")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}
	errObj := resp["error"].(map[string]any)
	if errObj["code"] != "invalid_count" {
		t.Errorf("expected error code invalid_count, got: %v", errObj["code"])
	}
}

func TestImage_InvalidAspect(t *testing.T) {
	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "A cat", "-o", "out.png", "--aspect", "5:4")

	if err == nil {
		t.Fatal("expected error for invalid aspect ratio")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}
	errObj := resp["error"].(map[string]any)
	if errObj["code"] != "invalid_aspect" {
		t.Errorf("expected error code invalid_aspect, got: %v", errObj["code"])
	}
}

func TestImage_InvalidResponseFormat(t *testing.T) {
	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "A cat", "-o", "out.png", "--response-format", "json")

	if err == nil {
		t.Fatal("expected error for invalid response format")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}
	errObj := resp["error"].(map[string]any)
	if errObj["code"] != "invalid_response_format" {
		t.Errorf("expected error code invalid_response_format, got: %v", errObj["code"])
	}
}

func TestImage_UnsupportedFormat(t *testing.T) {
	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "A cat", "-o", "out.gif")

	if err == nil {
		t.Fatal("expected error for unsupported output format")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}
	errObj := resp["error"].(map[string]any)
	if errObj["code"] != "unsupported_format" {
		t.Errorf("expected error code unsupported_format, got: %v", errObj["code"])
	}
}

func TestImage_InvalidSizeOnlyWidth(t *testing.T) {
	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "A cat", "-o", "out.png", "--width", "1024")

	if err == nil {
		t.Fatal("expected error for width without height")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}
	errObj := resp["error"].(map[string]any)
	if errObj["code"] != "invalid_size" {
		t.Errorf("expected error code invalid_size, got: %v", errObj["code"])
	}
}

func TestImage_InvalidSizeOutOfRange(t *testing.T) {
	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "A cat", "-o", "out.png", "--width", "256", "--height", "256")

	if err == nil {
		t.Fatal("expected error for size out of range")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}
	errObj := resp["error"].(map[string]any)
	if errObj["code"] != "invalid_size" {
		t.Errorf("expected error code invalid_size, got: %v", errObj["code"])
	}
}

func TestImage_InvalidSizeNotMultipleOf8(t *testing.T) {
	cmd := newImageCmd()
	// 1001 is not divisible by 8
	_, stderr, err := executeCommand(cmd, "A cat", "-o", "out.png", "--width", "1001", "--height", "1001")

	if err == nil {
		t.Fatal("expected error for size not multiple of 8")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}
	errObj := resp["error"].(map[string]any)
	if errObj["code"] != "invalid_size" {
		t.Errorf("expected error code invalid_size, got: %v", errObj["code"])
	}
}

func TestImage_ImageNotFound(t *testing.T) {
	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "A cat", "-o", "out.png", "-m", "image-01-live", "-i", "nonexistent.png")

	if err == nil {
		t.Fatal("expected error for image file not found")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}
	errObj := resp["error"].(map[string]any)
	if errObj["code"] != "image_read_error" {
		t.Errorf("expected error code image_read_error, got: %v", errObj["code"])
	}
}

func TestImage_AllFlags(t *testing.T) {
	cmd := newImageCmd()
	flags := cmd.Flags()

	expectedFlags := []string{
		"output", "prompt-file", "image", "model", "aspect",
		"width", "height", "n", "response-format", "prompt-optimizer",
	}

	for _, name := range expectedFlags {
		if flags.Lookup(name) == nil {
			t.Errorf("expected flag --%s to exist", name)
		}
	}
}

func TestImage_ShortFlags(t *testing.T) {
	cmd := newImageCmd()
	flags := cmd.Flags()

	shortFlags := map[string]string{
		"o": "output",
		"i": "image",
		"m": "model",
		"n": "n",
	}

	for short, long := range shortFlags {
		flag := flags.Lookup(long)
		if flag == nil {
			t.Errorf("expected flag --%s to exist", long)
			continue
		}
		if flag.Shorthand != short {
			t.Errorf("expected flag --%s to have shorthand -%s, got -%s", long, short, flag.Shorthand)
		}
	}
}

func TestImage_DefaultValues(t *testing.T) {
	cmd := newImageCmd()
	flags := cmd.Flags()

	defaults := map[string]string{
		"model":           "image-01",
		"aspect":          "1:1",
		"n":               "1",
		"response-format": "base64",
	}

	for name, expected := range defaults {
		flag := flags.Lookup(name)
		if flag == nil {
			t.Errorf("expected flag --%s to exist", name)
			continue
		}
		if flag.DefValue != expected {
			t.Errorf("expected flag --%s default to be %s, got %s", name, expected, flag.DefValue)
		}
	}
}
