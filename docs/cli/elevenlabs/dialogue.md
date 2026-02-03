# ElevenLabs Dialogue Command

## Usage

```bash
rawgenai elevenlabs dialogue -i <input.json> -o <output.mp3> [options]
```

## Input Format

Provide a JSON array of text and voice ID pairs:

```json
[
  {"text": "Knock knock", "voice_id": "Rachel"},
  {"text": "Who's there?", "voice_id": "Josh"},
  {"text": "Banana", "voice_id": "Rachel"},
  {"text": "Banana who?", "voice_id": "Josh"}
]
```

Voice IDs can be:
- Voice names: `Rachel`, `Josh`, `Bella`, `Antoni`, etc.
- Voice IDs: `21m00Tcm4TlvDq8ikWAM`

Maximum 10 unique voices per request.

## Input Sources

| Priority | Source | Example |
|----------|--------|---------|
| 1 | --input flag | `rawgenai elevenlabs dialogue -i dialogue.json -o out.mp3` |
| 2 | stdin | `cat dialogue.json \| rawgenai elevenlabs dialogue -o out.mp3` |

## Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| --output | -o | string | - | Yes* | Output file path (.mp3) |
| --input | -i | string | - | No | Input JSON file with dialogue data |
| --model | -m | string | "eleven_v3" | No | Model: eleven_v3, eleven_multilingual_v2 |
| --format | -f | string | "mp3_44100_128" | No | Output format |
| --language | -l | string | - | No | Language code (ISO 639-1) |
| --stability | - | float | 0.5 | No | Voice stability (0.0-1.0) |
| --text-normalization | - | string | "auto" | No | Text normalization: auto, on, off |
| --seed | - | int | - | No | Random seed for deterministic generation |
| --speak | - | bool | false | No | Play audio after generation |

*Required unless --speak is used

## Output Format

### Success
```json
{
  "success": true,
  "file": "/path/to/output.mp3",
  "model": "eleven_v3",
  "segments": 4
}
```

### Error
```json
{
  "success": false,
  "error": {
    "code": "too_many_voices",
    "message": "maximum 10 unique voices allowed per request"
  }
}
```

## Error Codes

### CLI Errors
| Code | Description |
|------|-------------|
| missing_input | No dialogue input provided |
| missing_output | Output file not specified |
| invalid_input | Invalid JSON input |
| empty_input | Dialogue array is empty |
| empty_text | Text is empty at some index |
| too_many_voices | More than 10 unique voices |
| invalid_stability | Stability out of range (0.0-1.0) |
| invalid_text_normalization | Invalid value (must be auto, on, off) |
| missing_api_key | ELEVENLABS_API_KEY not set |

## Examples

```bash
# Using input file
rawgenai elevenlabs dialogue -i dialogue.json -o conversation.mp3

# Using stdin
echo '[{"text":"Hello","voice_id":"Rachel"},{"text":"Hi there","voice_id":"Josh"}]' | \
  rawgenai elevenlabs dialogue -o chat.mp3

# With language specification
rawgenai elevenlabs dialogue -i dialogue.json -o chat.mp3 -l en

# With custom model
rawgenai elevenlabs dialogue -i dialogue.json -o chat.mp3 -m eleven_multilingual_v2

# Play immediately
rawgenai elevenlabs dialogue -i dialogue.json --speak
```

## API Reference

- Endpoint: `POST https://api.elevenlabs.io/v1/text-to-dialogue`
- Auth: `xi-api-key` header
- Docs: https://elevenlabs.io/docs/api-reference/text-to-dialogue/convert
