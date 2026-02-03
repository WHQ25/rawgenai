# ElevenLabs Voice Commands

Design new voices, manage voice library, and stream voice previews.

## Commands

| Command | Description |
|---------|-------------|
| `voice list` | List available voices with filtering |
| `voice design` | Design a new voice from text description |
| `voice create` | Create a permanent voice from a preview |
| `voice preview` | Stream a voice preview by ID |

---

## voice list

List available voices with filtering and pagination.

### Usage

```bash
rawgenai elevenlabs voice list [options]
```

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --search | string | - | Search term (searches name, description, labels) |
| --voice-type | string | - | Filter by type: personal, community, default, workspace, non-default, saved |
| --category | string | - | Filter by category: premade, cloned, generated, professional |
| --page-size | int | 10 | Results per page (max 100) |
| --page-token | string | - | Page token for pagination |
| --sort | string | - | Sort by: created_at_unix, name |
| --sort-dir | string | - | Sort direction: asc, desc |
| --collection-id | string | - | Filter by collection ID |
| --voice-ids | strings | - | Lookup specific voice IDs (comma-separated, max 100) |
| --total-count | bool | true | Include total count in response |

### Output

```json
{
  "success": true,
  "voices": [
    {
      "voice_id": "21m00Tcm4TlvDq8ikWAM",
      "name": "Rachel",
      "category": "premade",
      "description": "A warm, friendly American female voice",
      "preview_url": "https://...",
      "labels": {"accent": "american", "gender": "female"}
    }
  ],
  "has_more": true,
  "total_count": 150,
  "next_page_token": "eyJ..."
}
```

### Examples

```bash
# List all voices
rawgenai elevenlabs voice list

# Search for British voices
rawgenai elevenlabs voice list --search "british"

# Filter by category
rawgenai elevenlabs voice list --category cloned

# Personal voices only
rawgenai elevenlabs voice list --voice-type personal

# Sort by name
rawgenai elevenlabs voice list --sort name --sort-dir asc
```

---

## voice design

Design a new voice from a text description. Returns voice previews with audio samples.

### Usage

```bash
rawgenai elevenlabs voice design [description] -o <output.mp3> [options]
```

### Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| --output | -o | string | - | No | Output file for first preview |
| --description | - | string | - | No | Voice description (alternative to arg) |
| --text | - | string | - | No | Preview text (100-1000 chars) |
| --auto-text | - | bool | false | No | Auto-generate preview text |
| --model | -m | string | "eleven_multilingual_ttv_v2" | No | Model |
| --format | -f | string | "mp3_44100_128" | No | Output format |
| --loudness | - | float | 0.5 | No | Volume level (-1 to 1) |
| --guidance-scale | - | float | 5.0 | No | Prompt adherence (1-20) |
| --seed | - | int | - | No | Random seed |
| --enhance | - | bool | false | No | AI-enhance description |
| --reference-audio | - | string | - | No | Reference audio (v3 only) |
| --prompt-strength | - | float | 0.5 | No | Prompt vs reference (v3 only) |
| --speak | - | bool | false | No | Play first preview |

### Models

| Model | Description |
|-------|-------------|
| eleven_multilingual_ttv_v2 | Multilingual voice design (default) |
| eleven_ttv_v3 | Latest model with reference audio support |

### Output

```json
{
  "success": true,
  "file": "/path/to/preview.mp3",
  "previews": [
    {
      "generated_voice_id": "abc123...",
      "duration_secs": 5.2,
      "language": "en"
    }
  ],
  "text": "The preview text that was spoken"
}
```

### Examples

```bash
# Basic design
rawgenai elevenlabs voice design "A warm, friendly female voice with British accent" -o preview.mp3

# With custom preview text
rawgenai elevenlabs voice design "Deep male narrator" --text "Hello, welcome to our story" -o narrator.mp3

# Auto-generate text
rawgenai elevenlabs voice design "Energetic young voice" --auto-text -o preview.mp3

# Using reference audio (v3)
rawgenai elevenlabs voice design "Similar to this voice" --reference-audio sample.mp3 -m eleven_ttv_v3 -o preview.mp3

# Play immediately
rawgenai elevenlabs voice design "Cheerful assistant" --speak
```

---

## voice create

Create a permanent voice from a generated preview.

### Usage

```bash
rawgenai elevenlabs voice create --name <name> --description <desc> --voice-id <id>
```

### Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| --name | -n | string | - | Yes | Name for the new voice |
| --description | -d | string | - | Yes | Voice description |
| --voice-id | - | string | - | Yes | Generated voice ID from design |
| --labels | - | string | - | No | JSON object with labels |

### Output

```json
{
  "success": true,
  "voice_id": "xyz789...",
  "name": "My Custom Voice",
  "description": "A warm narrator voice",
  "labels": {"accent": "british"}
}
```

### Examples

```bash
# Basic create
rawgenai elevenlabs voice create --name "My Voice" --description "A custom voice" --voice-id abc123

# With labels
rawgenai elevenlabs voice create -n "Narrator" -d "Deep narrator" --voice-id xyz789 --labels '{"accent":"british","gender":"male"}'
```

---

## voice preview

Stream and download a voice preview by its generated ID.

### Usage

```bash
rawgenai elevenlabs voice preview <generated_voice_id> -o <output.mp3>
```

### Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| --output | -o | string | - | Yes* | Output file path |
| --speak | - | bool | false | No | Play audio |

*Required unless --speak is used

### Output

```json
{
  "success": true,
  "file": "/path/to/preview.mp3",
  "voice_id": "abc123..."
}
```

### Examples

```bash
# Download preview
rawgenai elevenlabs voice preview abc123 -o preview.mp3

# Play preview
rawgenai elevenlabs voice preview abc123 --speak
```

---

## Workflow Example

```bash
# 1. Design a voice and get preview
rawgenai elevenlabs voice design "A calm, soothing female narrator" -o preview.mp3
# Output: {"previews": [{"generated_voice_id": "abc123..."}]}

# 2. Listen to preview
rawgenai elevenlabs voice preview abc123 --speak

# 3. If satisfied, create permanent voice
rawgenai elevenlabs voice create -n "Calm Narrator" -d "Soothing female voice" --voice-id abc123
# Output: {"voice_id": "xyz789..."}

# 4. Use the new voice for TTS
rawgenai elevenlabs tts "Hello world" -v xyz789 -o hello.mp3
```

## API Reference

- Design: `POST https://api.elevenlabs.io/v1/text-to-voice/design`
- Create: `POST https://api.elevenlabs.io/v1/text-to-voice`
- Preview: `GET https://api.elevenlabs.io/v1/text-to-voice/{id}/stream`
- Docs: https://elevenlabs.io/docs/api-reference/text-to-voice
