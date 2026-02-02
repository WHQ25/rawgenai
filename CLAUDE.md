# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build
go build -o rawgenai ./cmd/rawgenai

# Run tests
go test ./...

# Run tests with verbose output
go test ./... -v

# Run a single test
go test ./internal/cli/openai/... -run TestTTS_MissingText -v

# Run the CLI
./rawgenai openai tts "Hello" -o hello.mp3
```

## Architecture

This is a CLI tool (`rawgenai`) for AI agents to access raw AI capabilities (TTS, STT, Image, Video).

**Command structure**: `rawgenai <provider> <action> [options]`

- Providers: `openai`, `google`, `elevenlabs`, `luma`, `replicate`
- Each provider has its own subcommands (e.g., `openai tts`, `openai stt`, `openai image`)
- All output is JSON (no text mode)

**Key directories**:
- `cmd/rawgenai/` - Entry point
- `internal/cli/` - CLI commands (root.go + provider subdirectories)
- `internal/cli/<provider>/` - Provider-specific commands
- `docs/cli/<provider>/` - CLI design docs
- `docs/research/<provider>/` - API research docs (endpoints, parameters, constraints)

## Development Workflow

**Develop one provider's one capability at a time** (e.g., OpenAI TTS, then OpenAI STT).

**Flow for each command**:
1. **Research** - Read provider's API docs, save to `docs/research/<provider>/`
2. **CLI Design** - Define command structure, flags, defaults
3. **Write docs** - Create `docs/cli/<provider>/<action>.md` with usage, flags, output format, error codes
4. **Write tests** - TDD: write unit tests first (`*_test.go`)
5. **Write implementation** - Make tests pass
6. **Manual test** - Verify with real API call

## Documentation Structure

**Two types of docs**:
- `docs/research/<provider>/` - API research (endpoints, parameters, constraints)
- `docs/cli/<provider>/` - CLI design docs

Each CLI command has a design doc at `docs/cli/<provider>/<action>.md` containing:
- Usage and examples
- Flags table (flag, short, type, default, required, description)
- Provider-specific options (models, voices, etc.)
- Output JSON format
- Error codes (CLI errors, API errors, network errors)

## Testing

Unit tests should cover:
- Parameter validation (missing required flags, invalid values)
- Input sources (positional arg, --file, stdin)
- Parameter compatibility (e.g., --instructions only works with certain models)
- Flag registration and default values

Tests use `cmd.SetOut()`, `cmd.SetErr()`, `cmd.SetIn()` to capture output without calling real APIs.

## Output Format

All commands output JSON to stdout (success) or stderr (error):

```json
// Success
{"success": true, "file": "/path/to/output.mp3", ...}

// Error
{"success": false, "error": {"code": "error_code", "message": "..."}}
```

Commands set `SilenceErrors: true` and `SilenceUsage: true` to ensure clean JSON output.

## Provider Development Patterns

### CLI Flag Naming Conventions

| Pattern | Examples | Description |
|---------|----------|-------------|
| `--first-frame`, `--last-frame` | Multi-word flags | Use kebab-case |
| `-i`, `-o`, `-v` | `--image`, `--output`, `--video` | Short flags for common inputs |
| `-m`, `-d`, `-r` | `--model`, `--duration`, `--ratio` | Short flags for common options |
| `-f` | `--prompt-file` | Read prompt from file |
| `--type`, `-t` | Task type selector | For commands with multiple API endpoints |
| `--verbose`, `-v` | Show full output | For hiding verbose info (e.g., URLs) by default |

### Image/Video Input Handling

Support both local files and URLs. Check API docs for base64 format requirements:

```go
// Some APIs require pure base64 (no data URL prefix)
func resolveImageURL(input string) (string, error) {
    if isURL(input) {
        return input, nil
    }
    return encodeImageToBase64(input)
}

func isURL(s string) bool {
    return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}
```

### Unit Test Coverage

Tests should cover these categories:

1. **Required field validation**
   - `TestXxx_MissingImage`, `TestXxx_MissingPrompt`

2. **File existence checks**
   - `TestXxx_ImageNotFound` - local file not found
   - URL inputs should skip file existence checks

3. **Invalid parameter values**
   - `TestXxx_InvalidModel`, `TestXxx_InvalidMode`

4. **Compatibility/conflict detection**
   - `TestXxx_IncompatibleXxx` - feature not supported by model/mode combination
   - `TestXxx_Compatible` - verify valid combinations pass validation

5. **Flag registration**
   - `TestXxx_AllFlags` - all expected flags exist
   - `TestXxx_ShortFlags` - short flag mappings correct
   - `TestXxx_DefaultValues` - default values correct

6. **JSON input parsing**
   - `TestXxx_InvalidXxxJSON` - malformed JSON input for complex flags

### Compatibility Check Pattern

When a feature is only supported by specific model/mode combinations:

```go
// Validate in order: model → mode → other constraints
if flags.someFeature != "" {
    if !supportedModels[flags.model] {
        return common.WriteError(cmd, "incompatible_some_feature",
            "some feature only supported by model X")
    }
    if flags.mode != "pro" {
        return common.WriteError(cmd, "incompatible_some_feature",
            "some feature only supported in pro mode")
    }
}
```

### Error Code Naming

| Pattern | Examples |
|---------|----------|
| `missing_xxx` | `missing_image`, `missing_api_key`, `missing_prompt` |
| `invalid_xxx` | `invalid_model`, `invalid_mode`, `invalid_format` |
| `xxx_not_found` | `image_not_found`, `file_not_found` |
| `xxx_read_error` | `image_read_error`, `file_read_error` |
| `incompatible_xxx` | `incompatible_feature`, `incompatible_mode` |

### Async Task Commands

For providers with async task APIs, create these commands:
- `create` / `generate` - submit task, return task_id
- `status` - query task status by task_id
- `download` - download completed result
- `list` - list tasks with pagination

When adding new task types, update `--type` flag in status/download/list commands.

### CLI Documentation Structure

Split large CLI docs (>500 lines) into focused files at `docs/cli/<provider>/`:
- `<capability>.md` - Overview, auth, error codes, command index
- `<capability>-<command>.md` - Individual command details
- `<capability>-examples.md` - Usage examples

### Research Documentation

Before implementing, create research docs at `docs/research/<provider>/`:
- API endpoints, request/response formats
- Parameter constraints and validation rules
- Model/feature compatibility matrix
