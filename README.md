# rawgenai

A CLI tool designed for AI agents to access raw generative AI capabilities.

## Design Philosophy

### Agent-First Design

This CLI is designed to be used by AI agents, not humans. Key design decisions:

- **JSON-only output**: All commands output structured JSON, making it easy for agents to parse results
- **Predictable error handling**: Errors are returned as structured JSON with consistent error codes
- **Minimal interactivity**: No prompts, confirmations, or interactive menus - pure input/output
- **Explicit parameters**: No magic defaults that might surprise an agent

### Provider-Centric Architecture

Different AI providers have different capabilities and parameters. Instead of hiding this complexity behind a unified interface, we expose it directly:

```
rawgenai <provider> <command> [subcommand] [options]
```

This design:
- Allows each provider to expose its unique features
- Avoids "lowest common denominator" abstractions
- Makes it clear which provider is being used
- Simplifies adding new providers

## Installation

**Homebrew (macOS):**
```bash
brew install WHQ25/tap/rawgenai
```

**Script (macOS/Linux):**
```bash
curl -fsSL https://raw.githubusercontent.com/WHQ25/rawgenai/main/install.sh | bash
```

**Binary Download:**
Download from [GitHub Releases](https://github.com/WHQ25/rawgenai/releases)

**From Source:**
```bash
go install github.com/WHQ25/rawgenai/cmd/rawgenai@latest
```

## Capabilities

| Category | Commands |
|----------|----------|
| **Image** | image |
| **Audio** | tts, stt, sfx, music, dialogue, voice |
| **Video** | video |

Notes:
- Commands vary by provider (some providers group audio features under `audio`)
- Many commands expose subcommands like `create`, `status`, `download`, `list`, `delete`

## Supported Providers & Commands

| Provider | Image | Audio | Video |
|----------|-------|-------|-------|
| OpenAI | image | tts, stt | video (create, remix, list, status, download, delete) |
| Google | image | tts, stt | video (create, extend, status, download) |
| ElevenLabs | - | tts, stt, sfx, music, dialogue, voices (list), voice (design, create, preview) | - |
| Grok | image | - | video (create, edit, status, download) |
| Seed | image | tts | video (create, status, download, list, delete) |
| Kling | image (create, list, status, download) | tts, voice (create, status, list, delete) | video (create, create-from-text, create-from-image, create-motion-control, create-avatar, extend, add-sound, status, download, list, element (create, list, delete)) |
| Runway | image (create, status, download, delete) | audio (tts, sfx, sts, dubbing, isolation, status, download, delete) | video (text2video, image2video, video2video, upscale, character, status, download, delete) |
| Luma | image (create, reframe, status, download, delete) | - | video (create, extend, upscale, audio, modify, list, status, download, delete) |
| MiniMax | image (create) | tts (create, status, download), music (create), voice (upload, clone, design, list, delete) | video (create, status, download) |
| DashScope | image | tts, stt (default, create, status) | video (create, status, download) |

## Documentation

- Provider CLI docs live under `docs/cli/` (where available)
- Use `rawgenai --help`, `rawgenai <provider> --help`, and `rawgenai <provider> <command> --help` to discover flags/subcommands

## CLI Structure

```
rawgenai <provider> <command> [subcommand] [options]
```

Providers: `openai`, `google`, `elevenlabs`, `grok`, `seed`, `kling`, `runway`, `luma`, `minimax`, `dashscope`

Commands vary by provider; common examples include `image`, `video`, `tts`, `stt`, and `audio`.

## Output Format

All output is JSON.

**Success:**
```json
{
  "success": true,
  ...
}
```

**Error:**
```json
{
  "success": false,
  "error": {
    "code": "error_code",
    "message": "Human readable message"
  }
}
```

## Configuration

**Priority**: CLI flags > Environment variables > Config file > Defaults

### Config Command

```bash
# Set API key
rawgenai config set openai_api_key sk-xxx

# List all config values
rawgenai config list

# Remove a config value
rawgenai config unset openai_api_key

# Show config file path
rawgenai config path
```

Config file: `~/.config/rawgenai/config.json`

### Environment Variables

- `OPENAI_API_KEY` - OpenAI
- `GEMINI_API_KEY` / `GOOGLE_API_KEY` - Google Gemini
- `ELEVENLABS_API_KEY` - ElevenLabs
- `XAI_API_KEY` - Grok
- `SEED_APP_ID`, `SEED_ACCESS_TOKEN` - ByteDance Seed TTS
- `ARK_API_KEY` - ByteDance Ark (Seed Image/Video)
- `KLING_ACCESS_KEY`, `KLING_SECRET_KEY` - Kling AI
- `KLING_BASE_URL` - Kling region/base URL (optional)
- `RUNWAY_API_KEY` - Runway
- `LUMA_API_KEY` - Luma AI
- `MINIMAX_API_KEY` - MiniMax
- `DASHSCOPE_API_KEY` - DashScope (Tongyi)
- `DASHSCOPE_BASE_URL` - DashScope base URL/region (optional)

## License

MIT
