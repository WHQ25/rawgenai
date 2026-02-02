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
rawgenai <provider> <action> [options]
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

| Capability | Description |
|------------|-------------|
| **TTS** | Text to Speech - Convert text to audio |
| **STT** | Speech to Text - Transcribe audio to text |
| **Image** | Generate images from text prompts |
| **Video** | Generate videos from text/image prompts |
| **Music** | Generate music from lyrics |

## Supported Providers

| Provider | TTS | STT | Image | Video | Music |
|----------|:---:|:---:|:-----:|:-----:|:-----:|
| OpenAI | ✅ | ✅ | ✅ | ✅ | - |
| Google | ✅ | ✅ | ✅ | ✅ | - |
| ElevenLabs | ✅ | ✅ | - | - | - |
| Grok | - | - | ✅ | ✅ | - |
| Seed | ✅ | - | ✅ | ✅ | - |
| Kling | - | - | - | ✅ | - |
| Runway | - | - | ✅ | ✅ | - |
| Luma | - | - | ✅ | ✅ | - |
| MiniMax | ✅ | - | ✅ | ✅ | ✅ |

## CLI Structure

```
rawgenai <provider> <action> [options]
```

Providers: `openai`, `google`, `elevenlabs`, `grok`, `seed`, `kling`, `runway`, `luma`, `minimax`

Actions: `tts`, `stt`, `image`, `video`, `music` (varies by provider)

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
- `GEMINI_API_KEY` - Google Gemini
- `ELEVENLABS_API_KEY` - ElevenLabs
- `XAI_API_KEY` - Grok
- `SEED_APP_ID`, `SEED_ACCESS_TOKEN` - ByteDance Seed TTS
- `ARK_API_KEY` - ByteDance Ark (Seed Image/Video)
- `KLING_ACCESS_KEY`, `KLING_SECRET_KEY` - Kling AI
- `RUNWAY_API_KEY` - Runway
- `LUMA_API_KEY` - Luma AI
- `MINIMAX_API_KEY` - MiniMax

## License

MIT
