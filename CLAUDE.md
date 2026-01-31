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
- `internal/cli/openai/` - OpenAI provider commands
- `docs/openai/` - CLI design docs per command (e.g., `tts.md`)

## Development Workflow

**Develop one provider's one capability at a time** (e.g., OpenAI TTS, then OpenAI STT).

**Flow for each command**:
1. **CLI Design** - Define command structure, flags, defaults
2. **Research** - Read provider's API docs, understand parameters and errors
3. **Write docs** - Create `docs/<provider>/<action>.md` with usage, flags, output format, error codes
4. **Write tests** - TDD: write unit tests first (`*_test.go`)
5. **Write implementation** - Make tests pass
6. **Manual test** - Verify with real API call

## Documentation Structure

Each command has a design doc at `docs/<provider>/<action>.md` containing:
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
