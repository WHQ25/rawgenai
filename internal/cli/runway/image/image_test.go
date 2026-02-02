package image

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCreate_MissingPrompt(t *testing.T) {
	imgFile := createTempFile(t, "ref.jpg", "fake image")

	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "create", "--ref-image", imgFile)

	if err == nil {
		t.Fatal("expected error for missing prompt")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_prompt" {
		t.Errorf("expected error code 'missing_prompt', got: %s", errorObj["code"])
	}
}

func TestCreate_MissingRefImage(t *testing.T) {
	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "create", "test prompt")

	if err == nil {
		t.Fatal("expected error for missing ref-image")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_ref_image" {
		t.Errorf("expected error code 'missing_ref_image', got: %s", errorObj["code"])
	}
}

func TestCreate_TooManyRefImages(t *testing.T) {
	img1 := createTempFile(t, "ref1.jpg", "fake image")
	img2 := createTempFile(t, "ref2.jpg", "fake image")
	img3 := createTempFile(t, "ref3.jpg", "fake image")
	img4 := createTempFile(t, "ref4.jpg", "fake image")

	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "create", "test",
		"--ref-image", img1, "--ref-image", img2,
		"--ref-image", img3, "--ref-image", img4)

	if err == nil {
		t.Fatal("expected error for too many ref-images")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "too_many_ref_images" {
		t.Errorf("expected error code 'too_many_ref_images', got: %s", errorObj["code"])
	}
}

func TestCreate_InvalidModel(t *testing.T) {
	imgFile := createTempFile(t, "ref.jpg", "fake image")

	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "create", "test", "--ref-image", imgFile, "-m", "invalid")

	if err == nil {
		t.Fatal("expected error for invalid model")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_model" {
		t.Errorf("expected error code 'invalid_model', got: %s", errorObj["code"])
	}
}

func TestCreate_InvalidRatio(t *testing.T) {
	imgFile := createTempFile(t, "ref.jpg", "fake image")

	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "create", "test", "--ref-image", imgFile, "-r", "invalid:ratio")

	if err == nil {
		t.Fatal("expected error for invalid ratio")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_ratio" {
		t.Errorf("expected error code 'invalid_ratio', got: %s", errorObj["code"])
	}
}

func TestCreate_ImageNotFound(t *testing.T) {
	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "create", "test", "--ref-image", "/nonexistent/image.jpg")

	if err == nil {
		t.Fatal("expected error for image not found")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "image_not_found" {
		t.Errorf("expected error code 'image_not_found', got: %s", errorObj["code"])
	}
}

func TestCreate_MissingAPIKey(t *testing.T) {
	setupNoConfigEnv(t)
	imgFile := createTempFile(t, "ref.jpg", "fake image")

	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "create", "test", "--ref-image", imgFile)

	if err == nil {
		t.Fatal("expected error for missing API key")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key', got: %s", errorObj["code"])
	}
}

func TestCreate_AllFlags(t *testing.T) {
	cmd := newCreateCmd()

	expectedFlags := []string{"ref-image", "ref-tag", "model", "ratio", "seed", "prompt-file", "public-figure"}
	for _, flag := range expectedFlags {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("expected flag '%s' not found", flag)
		}
	}
}

func TestCreate_DefaultValues(t *testing.T) {
	cmd := newCreateCmd()

	defaults := map[string]string{
		"model":         "gen4_image_turbo",
		"ratio":         "1024:1024",
		"seed":          "-1",
		"public-figure": "auto",
	}

	for flag, expected := range defaults {
		f := cmd.Flags().Lookup(flag)
		if f == nil {
			t.Errorf("flag '%s' not found", flag)
			continue
		}
		if f.DefValue != expected {
			t.Errorf("flag '%s' default is '%s', expected '%s'", flag, f.DefValue, expected)
		}
	}
}

// Status Tests
func TestStatus_MissingTaskID(t *testing.T) {
	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "status")

	if err == nil {
		t.Fatal("expected error for missing task_id")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_task_id" {
		t.Errorf("expected error code 'missing_task_id', got: %s", errorObj["code"])
	}
}

// Download Tests
func TestDownload_MissingTaskID(t *testing.T) {
	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "download", "-o", "output.png")

	if err == nil {
		t.Fatal("expected error for missing task_id")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_task_id" {
		t.Errorf("expected error code 'missing_task_id', got: %s", errorObj["code"])
	}
}

func TestDownload_MissingOutput(t *testing.T) {
	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "download", "test-task-id")

	if err == nil {
		t.Fatal("expected error for missing output")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_output" {
		t.Errorf("expected error code 'missing_output', got: %s", errorObj["code"])
	}
}

func TestDownload_InvalidOutputExtension(t *testing.T) {
	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "download", "test-task-id", "-o", "output.mp4")

	if err == nil {
		t.Fatal("expected error for invalid output extension")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_output" {
		t.Errorf("expected error code 'invalid_output', got: %s", errorObj["code"])
	}
}

// Delete Tests
func TestDelete_MissingTaskID(t *testing.T) {
	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "delete")

	if err == nil {
		t.Fatal("expected error for missing task_id")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_task_id" {
		t.Errorf("expected error code 'missing_task_id', got: %s", errorObj["code"])
	}
}
