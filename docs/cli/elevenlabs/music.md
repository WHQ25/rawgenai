# ElevenLabs Music Command

## Usage

```bash
rawgenai elevenlabs music [prompt] -o <output.mp3> [options]
```

## Input Sources

| Priority | Source | Example |
|----------|--------|---------|
| 1 | Positional argument | `rawgenai elevenlabs music "upbeat dance track" -o out.mp3` |
| 2 | --file flag | `rawgenai elevenlabs music --file prompt.txt -o out.mp3` |
| 3 | stdin | `echo "jazz fusion" \| rawgenai elevenlabs music -o out.mp3` |
| 4 | --composition-plan | `rawgenai elevenlabs music --composition-plan plan.json -o out.mp3` |

## Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| --output | -o | string | - | Yes* | Output file path (.mp3) |
| --duration | -d | int | - | No | Music length in milliseconds (3000-600000) |
| --instrumental | - | bool | false | No | Force instrumental (no vocals) |
| --format | -f | string | "mp3_44100_128" | No | Output format |
| --composition-plan | - | string | - | No | JSON file with detailed composition plan |
| --respect-durations | - | bool | true | No | Strictly respect section durations |
| --file | - | string | - | No | Input prompt file |
| --speak | - | bool | false | No | Play audio after generation |

*Required unless --speak is used

## Composition Plan

For detailed control, provide a JSON composition plan:

```json
{
  "positive_global_styles": ["electronic", "energetic"],
  "negative_global_styles": ["slow", "acoustic"],
  "sections": [
    {
      "section_name": "intro",
      "positive_local_styles": ["building", "atmospheric"],
      "negative_local_styles": ["loud"],
      "duration_ms": 15000,
      "lines": []
    },
    {
      "section_name": "verse",
      "positive_local_styles": ["rhythmic"],
      "negative_local_styles": [],
      "duration_ms": 30000,
      "lines": ["First verse lyrics here", "Second line of verse"]
    }
  ]
}
```

## Output Format

### Success
```json
{
  "success": true,
  "file": "/path/to/output.mp3",
  "duration_ms": 60000,
  "instrumental": false
}
```

### Error
```json
{
  "success": false,
  "error": {
    "code": "invalid_duration",
    "message": "duration must be between 3000 and 600000 milliseconds"
  }
}
```

## Error Codes

### CLI Errors
| Code | Description |
|------|-------------|
| missing_prompt | No prompt provided |
| missing_output | Output file not specified |
| invalid_duration | Duration out of range (3000-600000) |
| invalid_format | Unsupported output format |
| invalid_composition_plan | Invalid JSON in composition plan file |
| file_read_error | Cannot read input file |
| missing_api_key | ELEVENLABS_API_KEY not set |

## Examples

```bash
# Simple prompt
rawgenai elevenlabs music "upbeat electronic dance track" -o dance.mp3

# With duration
rawgenai elevenlabs music "calm piano melody" -o piano.mp3 -d 60000

# Instrumental only
rawgenai elevenlabs music "rock anthem with guitar solo" -o rock.mp3 --instrumental

# Using composition plan
rawgenai elevenlabs music --composition-plan song.json -o song.mp3

# Play immediately
rawgenai elevenlabs music "jazz fusion" --speak
```

## API Reference

- Endpoint: `POST https://api.elevenlabs.io/v1/music`
- Auth: `xi-api-key` header
- Docs: https://elevenlabs.io/docs/api-reference/music/compose
