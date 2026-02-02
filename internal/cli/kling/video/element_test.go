package video

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/WHQ25/rawgenai/internal/cli/common"
)

// ===== Element Create Command Tests =====

func TestElementCreate_MissingName(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "element", "create")

	if err == nil {
		t.Fatal("expected error for missing name")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_name" {
		t.Errorf("expected error code 'missing_name', got: %s", errorObj["code"])
	}
}

func TestElementCreate_MissingDescription(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "element", "create", "TestElement")

	if err == nil {
		t.Fatal("expected error for missing description")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_description" {
		t.Errorf("expected error code 'missing_description', got: %s", errorObj["code"])
	}
}

func TestElementCreate_MissingFrontal(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "element", "create", "TestElement", "-d", "Test description")

	if err == nil {
		t.Fatal("expected error for missing frontal image")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_frontal" {
		t.Errorf("expected error code 'missing_frontal', got: %s", errorObj["code"])
	}
}

func TestElementCreate_FrontalNotFound(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "element", "create", "TestElement",
		"-d", "Test description",
		"-f", "/nonexistent/frontal.jpg")

	if err == nil {
		t.Fatal("expected error for frontal image not found")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "frontal_not_found" {
		t.Errorf("expected error code 'frontal_not_found', got: %s", errorObj["code"])
	}
}

// Test that URL inputs skip file existence check for element
func TestElementCreate_FrontalURL(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("KLING_ACCESS_KEY", "")
	t.Setenv("KLING_SECRET_KEY", "")

	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "element", "create", "TestElement",
		"-d", "Test description",
		"-f", "https://example.com/frontal.jpg",
		"-r", "https://example.com/side.jpg")

	if err == nil {
		t.Fatal("expected error (missing api key)")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	// Should fail at API key check, not file check
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected URL to skip file check, got error: %s", errorObj["code"])
	}
}

func TestElementCreate_InvalidRefCount(t *testing.T) {
	frontalFile, err := os.CreateTemp("", "frontal_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(frontalFile.Name())
	frontalFile.Close()

	cmd := NewCmd()
	_, stderr, cmdErr := executeCommand(cmd, "element", "create", "TestElement",
		"-d", "Test description",
		"-f", frontalFile.Name())

	if cmdErr == nil {
		t.Fatal("expected error for missing ref images")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_ref_count" {
		t.Errorf("expected error code 'invalid_ref_count', got: %s", errorObj["code"])
	}
}

func TestElementCreate_RefNotFound(t *testing.T) {
	frontalFile, err := os.CreateTemp("", "frontal_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(frontalFile.Name())
	frontalFile.Close()

	cmd := NewCmd()
	_, stderr, cmdErr := executeCommand(cmd, "element", "create", "TestElement",
		"-d", "Test description",
		"-f", frontalFile.Name(),
		"-r", "/nonexistent/ref.jpg")

	if cmdErr == nil {
		t.Fatal("expected error for ref image not found")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "ref_not_found" {
		t.Errorf("expected error code 'ref_not_found', got: %s", errorObj["code"])
	}
}

func TestElementCreate_InvalidTag(t *testing.T) {
	frontalFile, err := os.CreateTemp("", "frontal_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(frontalFile.Name())
	frontalFile.Close()

	refFile, err := os.CreateTemp("", "ref_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(refFile.Name())
	refFile.Close()

	cmd := NewCmd()
	_, stderr, cmdErr := executeCommand(cmd, "element", "create", "TestElement",
		"-d", "Test description",
		"-f", frontalFile.Name(),
		"-r", refFile.Name(),
		"-t", "invalid_tag")

	if cmdErr == nil {
		t.Fatal("expected error for invalid tag")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_tag" {
		t.Errorf("expected error code 'invalid_tag', got: %s", errorObj["code"])
	}
}

func TestElementCreate_MissingAPIKey(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("KLING_ACCESS_KEY", "")
	t.Setenv("KLING_SECRET_KEY", "")

	frontalFile, err := os.CreateTemp("", "frontal_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(frontalFile.Name())
	frontalFile.Close()

	refFile, err := os.CreateTemp("", "ref_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(refFile.Name())
	refFile.Close()

	cmd := NewCmd()
	_, stderr, cmdErr := executeCommand(cmd, "element", "create", "TestElement",
		"-d", "Test description",
		"-f", frontalFile.Name(),
		"-r", refFile.Name())

	if cmdErr == nil {
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

func TestElementCreate_NameTooLong(t *testing.T) {
	cmd := NewCmd()
	longName := "ThisNameIsWayTooLongForAnElement"
	_, stderr, err := executeCommand(cmd, "element", "create", longName)

	if err == nil {
		t.Fatal("expected error for name too long")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_name" {
		t.Errorf("expected error code 'invalid_name', got: %s", errorObj["code"])
	}
}

func TestElementCreate_AllFlags(t *testing.T) {
	cmd := newElementCreateCmd()

	flags := []string{"description", "frontal", "ref", "tag"}
	for _, flag := range flags {
		if cmd.Flag(flag) == nil {
			t.Errorf("expected --%s flag", flag)
		}
	}
}

func TestElementCreate_ShortFlags(t *testing.T) {
	cmd := newElementCreateCmd()

	shortFlags := map[string]string{
		"d": "description",
		"f": "frontal",
		"r": "ref",
		"t": "tag",
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

// ===== Element List Command Tests =====

func TestElementList_MissingAPIKey(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("KLING_ACCESS_KEY", "")
	t.Setenv("KLING_SECRET_KEY", "")

	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "element", "list")

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

func TestElementList_InvalidType(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "element", "list", "--type", "invalid")

	if err == nil {
		t.Fatal("expected error for invalid type")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_type" {
		t.Errorf("expected error code 'invalid_type', got: %s", errorObj["code"])
	}
}

func TestElementList_InvalidLimit(t *testing.T) {
	tests := []struct {
		name  string
		limit string
	}{
		{"zero", "0"},
		{"negative", "-1"},
		{"too_large", "501"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewCmd()
			_, stderr, err := executeCommand(cmd, "element", "list", "--limit", tt.limit)

			if err == nil {
				t.Fatal("expected error for invalid limit")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "invalid_limit" {
				t.Errorf("expected error code 'invalid_limit', got: %s", errorObj["code"])
			}
		})
	}
}

func TestElementList_ValidTypes(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("KLING_ACCESS_KEY", "")
	t.Setenv("KLING_SECRET_KEY", "")

	types := []string{"custom", "official"}
	for _, tp := range types {
		t.Run(tp, func(t *testing.T) {
			cmd := NewCmd()
			_, stderr, err := executeCommand(cmd, "element", "list", "--type", tp)

			if err == nil {
				t.Fatal("expected error (missing api key)")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "missing_api_key" {
				t.Errorf("expected type '%s' to be valid, got error: %s", tp, errorObj["code"])
			}
		})
	}
}

func TestElementList_DefaultValues(t *testing.T) {
	cmd := newElementListCmd()

	if cmd.Flag("type").DefValue != "custom" {
		t.Errorf("expected default type 'custom', got: %s", cmd.Flag("type").DefValue)
	}
	if cmd.Flag("limit").DefValue != "30" {
		t.Errorf("expected default limit '30', got: %s", cmd.Flag("limit").DefValue)
	}
	if cmd.Flag("page").DefValue != "1" {
		t.Errorf("expected default page '1', got: %s", cmd.Flag("page").DefValue)
	}
}

// ===== Element Delete Command Tests =====

func TestElementDelete_MissingElementID(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "element", "delete")

	if err == nil {
		t.Fatal("expected error for missing element ID")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_element_id" {
		t.Errorf("expected error code 'missing_element_id', got: %s", errorObj["code"])
	}
}

func TestElementDelete_MissingAPIKey(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("KLING_ACCESS_KEY", "")
	t.Setenv("KLING_SECRET_KEY", "")

	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "element", "delete", "123456")

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
