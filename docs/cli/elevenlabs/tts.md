# ElevenLabs TTS Command

## Usage

```bash
rawgenai elevenlabs tts [text] -o <output.mp3> [options]
```

## Input Sources

| Priority | Source | Example |
|----------|--------|---------|
| 1 | Positional argument | `rawgenai elevenlabs tts "Hello world" -o out.mp3` |
| 2 | --file flag | `rawgenai elevenlabs tts --file input.txt -o out.mp3` |
| 3 | stdin | `echo "Hello" \| rawgenai elevenlabs tts -o out.mp3` |

## Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| --output | -o | string | - | Yes | Output file path (.mp3, .wav, .pcm) |
| --voice | -v | string | "Rachel" | No | Voice name or ID |
| --model | -m | string | "eleven_multilingual_v2" | No | Model ID |
| --format | -f | string | "mp3_44100_128" | No | Output format |
| --stability | - | float | 0.5 | No | Voice stability (0.0-1.0) |
| --similarity | - | float | 0.75 | No | Similarity boost (0.0-1.0) |
| --style | - | float | 0.0 | No | Style exaggeration (0.0-1.0) |
| --speed | - | float | 1.0 | No | Speaking speed (0.25-4.0) |
| --file | - | string | - | No | Input text file |

## Models

| Model ID | Description |
|----------|-------------|
| eleven_multilingual_v2 | Highest quality, 32 languages |
| eleven_flash_v2_5 | Low latency (~75ms), streaming |
| eleven_turbo_v2_5 | Fast, good quality |
| eleven_monolingual_v1 | English only, legacy |

## Output Formats

| Format | Description |
|--------|-------------|
| mp3_22050_32 | MP3, 22kHz, 32kbps |
| mp3_44100_128 | MP3, 44.1kHz, 128kbps (default) |
| mp3_44100_192 | MP3, 44.1kHz, 192kbps |
| pcm_16000 | PCM, 16kHz |
| pcm_44100 | PCM, 44.1kHz |

## Voices

Default voices (use name or ID):
- Rachel, Domi, Bella, Antoni, Elli, Josh, Arnold, Adam, Sam

Use `rawgenai elevenlabs voices` to list all available voices.

## Output Format

### Success
```json
{
  "success": true,
  "file": "/path/to/output.mp3",
  "voice": "Rachel",
  "model": "eleven_multilingual_v2",
  "characters": 42
}
```

### Error
```json
{
  "success": false,
  "error": {
    "code": "invalid_voice",
    "message": "Voice 'xyz' not found"
  }
}
```

## Error Codes

### CLI Errors
| Code | Description |
|------|-------------|
| missing_text | No text provided |
| missing_output | Output file not specified |
| invalid_speed | Speed out of range (0.25-4.0) |
| invalid_stability | Stability out of range (0.0-1.0) |
| invalid_similarity | Similarity out of range (0.0-1.0) |
| invalid_style | Style out of range (0.0-1.0) |
| invalid_format | Unsupported output format |
| missing_api_key | ELEVENLABS_API_KEY not set |

### API Errors (from ElevenLabs)
| Code | HTTP | Description |
|------|------|-------------|
| max_character_limit_exceeded | 400 | Text exceeds maximum character limit |
| invalid_api_key | 401 | API key is invalid or revoked |
| quota_exceeded | 401 | Character quota exhausted |
| voice_not_found | 400/404 | Voice ID not found |
| subscription_required | 403 | Professional voices require Creator+ plan |
| too_many_concurrent_requests | 429 | Exceeded concurrency limit |
| system_busy | 429 | ElevenLabs experiencing high traffic |
| rate_limit | 429 | Too many requests |
| server_error | 500 | ElevenLabs server error |
| server_overloaded | 503 | ElevenLabs server overloaded |

## Examples

```bash
# Basic usage
rawgenai elevenlabs tts "Hello world" -o hello.mp3

# With voice selection
rawgenai elevenlabs tts "Welcome" -o welcome.mp3 -v "Josh"

# High quality settings
rawgenai elevenlabs tts "Premium audio" -o premium.mp3 -f mp3_44100_192

# Adjust voice parameters
rawgenai elevenlabs tts "Calm narration" -o calm.mp3 --stability 0.8 --speed 0.9
```

## API Reference

- Endpoint: `POST https://api.elevenlabs.io/v1/text-to-speech/{voice_id}`
- Auth: `xi-api-key` header
- Docs: https://elevenlabs.io/docs/api-reference/text-to-speech/convert
