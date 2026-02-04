package image

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

func TestCreate_MissingPrompt(t *testing.T) {
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
	_, stderr, err := executeCommand(cmd, "create", "A cat", "--resolution", "bad")
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

func TestCreate_TooManyImages(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "create", "A cat",
		"-i", "https://a.com/1.png",
		"-i", "https://a.com/2.png",
		"-i", "https://a.com/3.png",
		"-i", "https://a.com/4.png")
	if err == nil {
		t.Fatal("expected error for too many images")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_image_count" {
		t.Errorf("expected error code 'invalid_image_count', got: %s", errorObj["code"])
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

func TestCreate_AllFlags(t *testing.T) {
	cmd := NewCmd()
	createCmd, _, _ := cmd.Find([]string{"create"})
	if createCmd == nil {
		t.Fatal("create command not found")
	}

	expectedFlags := []string{"image", "resolution", "seed", "no-revise", "no-watermark", "region", "prompt-file"}
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
		"resolution": "1024:1024",
		"seed":       "0",
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

func TestCreate_PromptFromStdin(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("TENCENT_SECRET_ID", "")
	t.Setenv("TENCENT_SECRET_KEY", "")

	cmd := NewCmd()
	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)
	cmd.SetOut(stdoutBuf)
	cmd.SetErr(stderrBuf)
	cmd.SetIn(strings.NewReader("A cat from stdin"))
	cmd.SetArgs([]string{"create"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error (missing API key), but no error")
	}

	// Should fail with missing_api_key, not missing_prompt
	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderrBuf.String())), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderrBuf.String())
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key', got: %s", errorObj["code"])
	}
}
