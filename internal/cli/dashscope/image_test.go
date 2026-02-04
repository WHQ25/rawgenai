package dashscope

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/WHQ25/rawgenai/internal/cli/common"
)

func executeImageCommand(args ...string) (stdout, stderr string, err error) {
	cmd := newImageCmd()
	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)

	cmd.SetOut(stdoutBuf)
	cmd.SetErr(stderrBuf)
	cmd.SetArgs(args)

	err = cmd.Execute()
	return stdoutBuf.String(), stderrBuf.String(), err
}

func expectImageErrorCode(t *testing.T, stderr string, expectedCode string) {
	t.Helper()
	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}
	if resp["success"] != false {
		t.Error("expected success to be false")
	}
	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != expectedCode {
		t.Errorf("expected error code '%s', got: %s", expectedCode, errorObj["code"])
	}
}

// ===== Prompt Validation =====

func TestImage_MissingPrompt(t *testing.T) {
	common.SetupNoConfigEnv(t)
	_, stderr, err := executeImageCommand()
	if err == nil {
		t.Fatal("expected error for missing prompt")
	}
	expectImageErrorCode(t, stderr, "missing_prompt")
}

func TestImage_FromFile(t *testing.T) {
	common.SetupNoConfigEnv(t)
	tmpFile, err := os.CreateTemp("", "prompt_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("a cat in a garden")
	tmpFile.Close()

	// Should pass prompt validation and fail at API key
	_, stderr, err := executeImageCommand("-f", tmpFile.Name())
	if err == nil {
		t.Fatal("expected error")
	}
	expectImageErrorCode(t, stderr, "missing_api_key")
}

func TestImage_FromStdin(t *testing.T) {
	common.SetupNoConfigEnv(t)
	cmd := newImageCmd()
	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)
	cmd.SetOut(stdoutBuf)
	cmd.SetErr(stderrBuf)
	cmd.SetArgs([]string{})
	cmd.SetIn(strings.NewReader("a cat in a garden"))

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	expectImageErrorCode(t, stderrBuf.String(), "missing_api_key")
}

// ===== Mode and Model Validation =====

func TestImage_InvalidModel(t *testing.T) {
	common.SetupNoConfigEnv(t)
	_, stderr, err := executeImageCommand("a cat", "-m", "invalid-model")
	if err == nil {
		t.Fatal("expected error for invalid model")
	}
	expectImageErrorCode(t, stderr, "invalid_model")
}

func TestImage_IncompatibleImageWithT2I(t *testing.T) {
	common.SetupNoConfigEnv(t)
	// T2I models should not accept --image
	t2iModels := []string{"wan2.6-t2i", "qwen-image-max", "qwen-image-plus"}
	for _, model := range t2iModels {
		t.Run(model, func(t *testing.T) {
			_, stderr, err := executeImageCommand("a cat", "-m", model, "-i", "https://example.com/img.jpg")
			if err == nil {
				t.Fatal("expected error")
			}
			expectImageErrorCode(t, stderr, "incompatible_image")
		})
	}
}

func TestImage_MissingImageForEdit(t *testing.T) {
	common.SetupNoConfigEnv(t)
	// Edit models require --image
	editModels := []string{"qwen-image-edit-max", "qwen-image-edit-plus", "qwen-image-edit", "wan2.6-image"}
	for _, model := range editModels {
		t.Run(model, func(t *testing.T) {
			_, stderr, err := executeImageCommand("edit this", "-m", model)
			if err == nil {
				t.Fatal("expected error")
			}
			expectImageErrorCode(t, stderr, "missing_image")
		})
	}
}

// ===== Image Input Validation =====

func TestImage_ImageNotFound(t *testing.T) {
	common.SetupNoConfigEnv(t)
	_, stderr, err := executeImageCommand("edit this", "-i", "/nonexistent/image.jpg")
	if err == nil {
		t.Fatal("expected error for missing image file")
	}
	expectImageErrorCode(t, stderr, "image_not_found")
}

func TestImage_ImageURL(t *testing.T) {
	common.SetupNoConfigEnv(t)
	// URL should skip file existence check and reach API key validation
	_, stderr, err := executeImageCommand("edit this", "-i", "https://example.com/photo.jpg")
	if err == nil {
		t.Fatal("expected error")
	}
	expectImageErrorCode(t, stderr, "missing_api_key")
}

func TestImage_TooManyImagesQwenEdit(t *testing.T) {
	common.SetupNoConfigEnv(t)
	// qwen-image-edit models accept max 3 images
	_, stderr, err := executeImageCommand("edit this",
		"-m", "qwen-image-edit-plus",
		"-i", "https://a.com/1.jpg",
		"-i", "https://a.com/2.jpg",
		"-i", "https://a.com/3.jpg",
		"-i", "https://a.com/4.jpg",
	)
	if err == nil {
		t.Fatal("expected error for too many images")
	}
	expectImageErrorCode(t, stderr, "too_many_images")
}

func TestImage_TooManyImagesWanImage(t *testing.T) {
	common.SetupNoConfigEnv(t)
	// wan2.6-image accepts max 4 images
	_, stderr, err := executeImageCommand("edit this",
		"-m", "wan2.6-image",
		"-i", "https://a.com/1.jpg",
		"-i", "https://a.com/2.jpg",
		"-i", "https://a.com/3.jpg",
		"-i", "https://a.com/4.jpg",
		"-i", "https://a.com/5.jpg",
	)
	if err == nil {
		t.Fatal("expected error for too many images")
	}
	expectImageErrorCode(t, stderr, "too_many_images")
}

func TestImage_MaxImagesQwenEdit(t *testing.T) {
	common.SetupNoConfigEnv(t)
	// Exactly 3 images should pass for qwen-edit
	_, stderr, err := executeImageCommand("edit this",
		"-m", "qwen-image-edit-plus",
		"-i", "https://a.com/1.jpg",
		"-i", "https://a.com/2.jpg",
		"-i", "https://a.com/3.jpg",
	)
	if err == nil {
		t.Fatal("expected error")
	}
	// Should pass image validation and reach API key
	expectImageErrorCode(t, stderr, "missing_api_key")
}

func TestImage_MaxImagesWanImage(t *testing.T) {
	common.SetupNoConfigEnv(t)
	// Exactly 4 images should pass for wan2.6-image
	_, stderr, err := executeImageCommand("edit this",
		"-m", "wan2.6-image",
		"-i", "https://a.com/1.jpg",
		"-i", "https://a.com/2.jpg",
		"-i", "https://a.com/3.jpg",
		"-i", "https://a.com/4.jpg",
	)
	if err == nil {
		t.Fatal("expected error")
	}
	expectImageErrorCode(t, stderr, "missing_api_key")
}

// ===== Count Validation =====

func TestImage_InvalidCount(t *testing.T) {
	common.SetupNoConfigEnv(t)
	_, stderr, err := executeImageCommand("a cat", "-n", "0")
	if err == nil {
		t.Fatal("expected error for invalid count")
	}
	expectImageErrorCode(t, stderr, "invalid_count")
}

func TestImage_InvalidCountQwenImage(t *testing.T) {
	common.SetupNoConfigEnv(t)
	// qwen-image models only support n=1
	_, stderr, err := executeImageCommand("a cat", "-m", "qwen-image-max", "-n", "2")
	if err == nil {
		t.Fatal("expected error for count > 1 with qwen-image")
	}
	expectImageErrorCode(t, stderr, "invalid_count")
}

func TestImage_InvalidCountQwenEditBasic(t *testing.T) {
	common.SetupNoConfigEnv(t)
	// qwen-image-edit (basic) only supports n=1
	_, stderr, err := executeImageCommand("edit", "-m", "qwen-image-edit", "-i", "https://a.com/1.jpg", "-n", "2")
	if err == nil {
		t.Fatal("expected error for count > 1 with qwen-image-edit")
	}
	expectImageErrorCode(t, stderr, "invalid_count")
}

func TestImage_ValidCountWan(t *testing.T) {
	common.SetupNoConfigEnv(t)
	// wan2.6-t2i supports n=1-4
	_, stderr, err := executeImageCommand("a cat", "-m", "wan2.6-t2i", "-n", "4")
	if err == nil {
		t.Fatal("expected error")
	}
	expectImageErrorCode(t, stderr, "missing_api_key")
}

func TestImage_InvalidCountWan(t *testing.T) {
	common.SetupNoConfigEnv(t)
	_, stderr, err := executeImageCommand("a cat", "-m", "wan2.6-t2i", "-n", "5")
	if err == nil {
		t.Fatal("expected error for count > 4 with wan model")
	}
	expectImageErrorCode(t, stderr, "invalid_count")
}

func TestImage_ValidCountQwenEditMax(t *testing.T) {
	common.SetupNoConfigEnv(t)
	// qwen-image-edit-max supports n=1-6
	_, stderr, err := executeImageCommand("edit", "-m", "qwen-image-edit-max", "-i", "https://a.com/1.jpg", "-n", "6")
	if err == nil {
		t.Fatal("expected error")
	}
	expectImageErrorCode(t, stderr, "missing_api_key")
}

func TestImage_InvalidCountQwenEditMax(t *testing.T) {
	common.SetupNoConfigEnv(t)
	_, stderr, err := executeImageCommand("edit", "-m", "qwen-image-edit-max", "-i", "https://a.com/1.jpg", "-n", "7")
	if err == nil {
		t.Fatal("expected error for count > 6 with qwen-edit-max")
	}
	expectImageErrorCode(t, stderr, "invalid_count")
}

// ===== Compatibility =====

func TestImage_IncompatiblePromptExtend(t *testing.T) {
	common.SetupNoConfigEnv(t)
	// qwen-image-edit (basic) does not support prompt-extend
	// prompt-extend defaults to true, so explicitly setting it should error for basic model
	_, stderr, err := executeImageCommand("edit", "-m", "qwen-image-edit", "-i", "https://a.com/1.jpg", "--prompt-extend=true")
	if err == nil {
		t.Fatal("expected error")
	}
	expectImageErrorCode(t, stderr, "incompatible_prompt_extend")
}

// ===== API Key =====

func TestImage_MissingAPIKey(t *testing.T) {
	common.SetupNoConfigEnv(t)
	_, stderr, err := executeImageCommand("a cat in a garden")
	if err == nil {
		t.Fatal("expected error for missing API key")
	}
	expectImageErrorCode(t, stderr, "missing_api_key")
}

// ===== Flag Registration =====

func TestImage_AllFlags(t *testing.T) {
	cmd := newImageCmd()
	expectedFlags := []string{
		"model", "image", "size", "count", "negative",
		"seed", "prompt-extend", "watermark", "output", "prompt-file",
	}
	for _, name := range expectedFlags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("missing flag: %s", name)
		}
	}
}

func TestImage_ShortFlags(t *testing.T) {
	cmd := newImageCmd()
	shortFlags := map[string]string{
		"model":       "m",
		"image":       "i",
		"size":        "s",
		"count":       "n",
		"output":      "o",
		"prompt-file": "f",
	}
	for name, short := range shortFlags {
		flag := cmd.Flags().Lookup(name)
		if flag == nil {
			t.Errorf("missing flag: %s", name)
			continue
		}
		if flag.Shorthand != short {
			t.Errorf("flag %s: expected short '%s', got '%s'", name, short, flag.Shorthand)
		}
	}
}

func TestImage_DefaultValues(t *testing.T) {
	cmd := newImageCmd()

	// model defaults to empty (auto-selected)
	if f := cmd.Flags().Lookup("model"); f.DefValue != "" {
		t.Errorf("model default should be empty, got: %s", f.DefValue)
	}

	// count defaults to 1
	if f := cmd.Flags().Lookup("count"); f.DefValue != "1" {
		t.Errorf("count default should be 1, got: %s", f.DefValue)
	}

	// prompt-extend defaults to true
	if f := cmd.Flags().Lookup("prompt-extend"); f.DefValue != "true" {
		t.Errorf("prompt-extend default should be true, got: %s", f.DefValue)
	}

	// watermark defaults to false
	if f := cmd.Flags().Lookup("watermark"); f.DefValue != "false" {
		t.Errorf("watermark default should be false, got: %s", f.DefValue)
	}
}
