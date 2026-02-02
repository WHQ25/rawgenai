package video

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestImage2Video_MissingImage(t *testing.T) {
	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "image2video")

	if err == nil {
		t.Fatal("expected error for missing image")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	if resp["success"] != false {
		t.Error("expected success to be false")
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_image" {
		t.Errorf("expected error code 'missing_image', got: %s", errorObj["code"])
	}
}

func TestImage2Video_InvalidModel(t *testing.T) {
	imgFile := createTempFile(t, "test.jpg", "fake image data")

	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "image2video", "-i", imgFile, "-m", "invalid_model")

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

func TestImage2Video_InvalidRatio(t *testing.T) {
	imgFile := createTempFile(t, "test.jpg", "fake image data")

	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "image2video", "-i", imgFile, "-r", "invalid:ratio")

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

func TestImage2Video_InvalidDuration(t *testing.T) {
	imgFile := createTempFile(t, "test.jpg", "fake image data")

	tests := []struct {
		name     string
		duration string
	}{
		{"too_short", "1"},
		{"too_long", "11"},
		{"zero", "0"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := newTestCmd()
			_, stderr, _ := executeCommand(cmd, "image2video", "-i", imgFile, "-d", test.duration)

			var resp map[string]any
			json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "invalid_duration" {
				t.Errorf("expected invalid_duration, got: %s", errorObj["code"])
			}
		})
	}
}

func TestImage2Video_InvalidPublicFigure(t *testing.T) {
	imgFile := createTempFile(t, "test.jpg", "fake image data")

	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "image2video", "-i", imgFile, "--public-figure", "invalid")

	if err == nil {
		t.Fatal("expected error for invalid public-figure")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_public_figure" {
		t.Errorf("expected error code 'invalid_public_figure', got: %s", errorObj["code"])
	}
}

func TestImage2Video_ImageNotFound(t *testing.T) {
	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "image2video", "-i", "/nonexistent/image.jpg")

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

func TestImage2Video_ImageURL_SkipsFileCheck(t *testing.T) {
	setupNoConfigEnv(t)

	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "image2video", "-i", "https://example.com/image.jpg")

	if err == nil {
		t.Fatal("expected error (missing API key), but should not be image_not_found")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	// Should fail on API key, not file check
	if errorObj["code"] == "image_not_found" {
		t.Error("URL should skip file existence check")
	}
}

func TestImage2Video_MissingAPIKey(t *testing.T) {
	setupNoConfigEnv(t)

	imgFile := createTempFile(t, "test.jpg", "fake image data")

	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "image2video", "-i", imgFile)

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

func TestImage2Video_AllFlags(t *testing.T) {
	cmd := newImage2VideoCmd()

	expectedFlags := []string{"image", "model", "ratio", "duration", "seed", "prompt-file", "public-figure"}
	for _, flag := range expectedFlags {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("expected flag '%s' not found", flag)
		}
	}
}

func TestImage2Video_ShortFlags(t *testing.T) {
	cmd := newImage2VideoCmd()

	shortFlags := map[string]string{
		"i": "image",
		"m": "model",
		"r": "ratio",
		"d": "duration",
		"f": "prompt-file",
	}

	for short, long := range shortFlags {
		flag := cmd.Flags().ShorthandLookup(short)
		if flag == nil {
			t.Errorf("expected short flag '-%s' not found", short)
			continue
		}
		if flag.Name != long {
			t.Errorf("short flag '-%s' maps to '%s', expected '%s'", short, flag.Name, long)
		}
	}
}

func TestImage2Video_DefaultValues(t *testing.T) {
	cmd := newImage2VideoCmd()

	defaults := map[string]string{
		"model":         "gen4_turbo",
		"ratio":         "1280:720",
		"duration":      "5",
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

func TestImage2Video_ValidModels(t *testing.T) {
	setupNoConfigEnv(t)

	validModels := []string{"gen4_turbo", "veo3.1", "veo3.1_fast", "gen3a_turbo", "veo3"}
	imgFile := createTempFile(t, "test.jpg", "fake image data")

	for _, model := range validModels {
		t.Run(model, func(t *testing.T) {
			cmd := newTestCmd()
			_, stderr, _ := executeCommand(cmd, "image2video", "-i", imgFile, "-m", model)

			var resp map[string]any
			json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

			errorObj, ok := resp["error"].(map[string]any)
			if ok && errorObj["code"] == "invalid_model" {
				t.Errorf("model '%s' should be valid", model)
			}
		})
	}
}

func TestImage2Video_ValidRatios(t *testing.T) {
	setupNoConfigEnv(t)

	validRatios := []string{"1280:720", "720:1280", "1104:832", "832:1104", "960:960", "1584:672"}
	imgFile := createTempFile(t, "test.jpg", "fake image data")

	for _, ratio := range validRatios {
		t.Run(ratio, func(t *testing.T) {
			cmd := newTestCmd()
			_, stderr, _ := executeCommand(cmd, "image2video", "-i", imgFile, "-r", ratio)

			var resp map[string]any
			json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

			errorObj, ok := resp["error"].(map[string]any)
			if ok && errorObj["code"] == "invalid_ratio" {
				t.Errorf("ratio '%s' should be valid", ratio)
			}
		})
	}
}

func TestImage2Video_ValidDurations(t *testing.T) {
	setupNoConfigEnv(t)

	validDurations := []string{"2", "5", "10"}
	imgFile := createTempFile(t, "test.jpg", "fake image data")

	for _, duration := range validDurations {
		t.Run(duration, func(t *testing.T) {
			cmd := newTestCmd()
			_, stderr, _ := executeCommand(cmd, "image2video", "-i", imgFile, "-d", duration)

			var resp map[string]any
			json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

			errorObj, ok := resp["error"].(map[string]any)
			if ok && errorObj["code"] == "invalid_duration" {
				t.Errorf("duration '%s' should be valid", duration)
			}
		})
	}
}
