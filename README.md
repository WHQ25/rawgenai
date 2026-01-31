# rawgenai

A CLI tool designed for AI agents to access raw AI capabilities.

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

## Capabilities

| Capability | Description |
|------------|-------------|
| **TTS** | Text to Speech - Convert text to audio |
| **STT** | Speech to Text - Transcribe audio to text |
| **Image** | Generate images from text prompts |
| **Video** | Generate videos from text/keyframes |

## Supported Providers

| Provider | TTS | STT | Image | Video |
|----------|:---:|:---:|:-----:|:-----:|
| OpenAI | ✅ | ✅ | ✅ | - |
| Google | ✅ | ✅ | ✅ | - |
| ElevenLabs | ✅ | - | - | - |
| Luma | - | - | - | ✅ |
| Replicate | - | - | - | ✅ |

## CLI Structure

```
rawgenai
├── openai
│   ├── tts      # Text to Speech
│   ├── stt      # Speech to Text
│   └── image    # Image generation
├── google
│   ├── tts      # Text to Speech
│   ├── stt      # Speech to Text
│   └── image    # Image generation
├── elevenlabs
│   └── tts      # Text to Speech
├── luma
│   └── video    # Video generation (sync + async)
├── replicate
│   └── video    # Video generation (sync + async)
└── config
    ├── init     # Initialize config file
    ├── set      # Set config value
    ├── get      # Get config value
    └── show     # Show all config
```

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

**Config file**: `~/.config/rawgenai/config.yaml`

**Environment variables**:
- `OPENAI_API_KEY`
- `ELEVENLABS_API_KEY`
- `LUMA_API_KEY`
- `REPLICATE_API_TOKEN`

## Project Structure

```
rawgenai/
├── cmd/rawgenai/
│   └── main.go
├── internal/
│   ├── cli/
│   │   ├── root.go
│   │   ├── config.go
│   │   ├── openai/
│   │   ├── elevenlabs/
│   │   ├── luma/
│   │   └── replicate/
│   ├── config/
│   └── output/
├── go.mod
└── README.md
```

## License

MIT
