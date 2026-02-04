package video

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/spf13/cobra"
)

func executeCommand(root *cobra.Command, args ...string) (stdout, stderr string, err error) {
	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)

	root.SetOut(stdoutBuf)
	root.SetErr(stderrBuf)
	root.SetArgs(args)

	err = root.Execute()
	return stdoutBuf.String(), stderrBuf.String(), err
}

func TestCreate_MissingPromptAndImage(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "create")
	if err == nil {
		t.Fatal("expected error for missing prompt")
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

func TestCreate_InvalidResolution(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "create", "A cat", "--resolution", "1080p")
	if err == nil {
		t.Fatal("expected error for invalid resolution")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_resolution" {
		t.Errorf("expected error code 'invalid_resolution', got: %s", errorObj["code"])
	}
}

func TestCreate_ImageNotFound(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "create", "A cat", "--image", "/nonexistent/image.png")
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

func TestCreate_ImageURLSkipsFileCheck(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("TENCENT_SECRET_ID", "")
	t.Setenv("TENCENT_SECRET_KEY", "")

	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "create", "A cat", "--image", "https://example.com/image.png")
	if err == nil {
		t.Fatal("expected error (missing API key)")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	// Should fail with missing_api_key, not image_not_found
	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key', got: %s", errorObj["code"])
	}
}

func TestCreate_MissingAPIKey(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("TENCENT_SECRET_ID", "")
	t.Setenv("TENCENT_SECRET_KEY", "")

	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "create", "A cat")
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

func TestCreate_ImageOnlyNoPrompt(t *testing.T) {
	// Image-to-video should work without prompt; should reach API key check
	common.SetupNoConfigEnv(t)
	t.Setenv("TENCENT_SECRET_ID", "")
	t.Setenv("TENCENT_SECRET_KEY", "")

	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "create", "--image", "https://example.com/image.png")
	if err == nil {
		t.Fatal("expected error (missing API key)")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key' (not missing_prompt), got: %s", errorObj["code"])
	}
}

func TestCreate_AllFlags(t *testing.T) {
	cmd := NewCmd()
	createCmd, _, _ := cmd.Find([]string{"create"})
	if createCmd == nil {
		t.Fatal("create command not found")
	}

	expectedFlags := []string{"image", "resolution", "no-watermark", "region", "prompt-file"}
	for _, name := range expectedFlags {
		if createCmd.Flags().Lookup(name) == nil {
			t.Errorf("expected flag --%s not found", name)
		}
	}
}

func TestCreate_ShortFlags(t *testing.T) {
	cmd := NewCmd()
	createCmd, _, _ := cmd.Find([]string{"create"})
	if createCmd == nil {
		t.Fatal("create command not found")
	}

	shortFlags := map[string]string{
		"i": "image",
		"r": "resolution",
		"f": "prompt-file",
	}

	for short, long := range shortFlags {
		flag := createCmd.Flags().Lookup(long)
		if flag == nil {
			t.Errorf("flag --%s not found", long)
			continue
		}
		if flag.Shorthand != short {
			t.Errorf("expected short flag -%s for --%s, got -%s", short, long, flag.Shorthand)
		}
	}
}

func TestCreate_DefaultValues(t *testing.T) {
	cmd := NewCmd()
	createCmd, _, _ := cmd.Find([]string{"create"})
	if createCmd == nil {
		t.Fatal("create command not found")
	}

	defaults := map[string]string{
		"resolution": "720p",
		"region":     "ap-guangzhou",
	}

	for name, expected := range defaults {
		flag := createCmd.Flags().Lookup(name)
		if flag == nil {
			t.Errorf("flag --%s not found", name)
			continue
		}
		if flag.DefValue != expected {
			t.Errorf("expected default %s for --%s, got %s", expected, name, flag.DefValue)
		}
	}
}
