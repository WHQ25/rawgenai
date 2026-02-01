# ElevenLabs Sound Effects Command

## Usage

```bash
rawgenai elevenlabs sfx [prompt] -o <output.mp3> [options]
```

## Input Sources

| Priority | Source | Example |
|----------|--------|---------|
| 1 | Positional argument | `rawgenai elevenlabs sfx "explosion" -o boom.mp3` |
| 2 | --prompt-file flag | `rawgenai elevenlabs sfx --prompt-file prompt.txt -o out.mp3` |
| 3 | stdin | `echo "thunder" \| rawgenai elevenlabs sfx -o thunder.mp3` |

## Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| --output | -o | string | - | Yes | Output file path (.mp3) |
| --duration | -d | float | auto | No | Duration in seconds (0.5-30) |
| --loop | - | bool | false | No | Create seamless loop |
| --influence | - | float | 0.3 | No | Prompt influence (0.0-1.0) |
| --format | -f | string | "mp3_44100_128" | No | Output format |
| --prompt-file | - | string | - | No | Input prompt file |

## Output Formats

| Format | Description |
|--------|-------------|
| mp3_22050_32 | MP3, 22kHz, 32kbps |
| mp3_44100_128 | MP3, 44.1kHz, 128kbps (default) |
| mp3_44100_192 | MP3, 44.1kHz, 192kbps |
| pcm_44100 | PCM, 44.1kHz |
| pcm_48000 | PCM, 48kHz |

## Output Format

### Success
```json
{
  "success": true,
  "file": "/path/to/output.mp3",
  "duration": 5.0,
  "loop": false
}
```

### Error
```json
{
  "success": false,
  "error": {
    "code": "invalid_duration",
    "message": "Duration must be between 0.5 and 30 seconds"
  }
}
```

## Error Codes

### CLI Errors
| Code | Description |
|------|-------------|
| missing_prompt | No prompt provided |
| missing_output | Output file not specified |
| invalid_duration | Duration out of range (0.5-30s) |
| invalid_influence | Influence out of range (0.0-1.0) |
| invalid_format | Unsupported output format |
| missing_api_key | ELEVENLABS_API_KEY not set |

### API Errors (from ElevenLabs)
| Code | HTTP | Description |
|------|------|-------------|
| invalid_api_key | 401 | API key is invalid or revoked |
| quota_exceeded | 401 | API quota exhausted |
| subscription_required | 403 | Feature requires higher plan |
| too_many_concurrent_requests | 429 | Exceeded concurrency limit |
| system_busy | 429 | ElevenLabs experiencing high traffic |
| rate_limit | 429 | Too many requests |
| server_error | 500 | ElevenLabs server error |

## Prompt Tips

**Good prompts:**
- "Heavy rain on a metal roof with distant thunder"
- "Footsteps on gravel, slow walking pace"
- "Sci-fi door whoosh, mechanical"
- "Crowd cheering in a stadium, excited"

**Tips:**
- Be specific about the sound characteristics
- Mention material, intensity, environment
- For complex sounds, generate components separately

## Examples

```bash
# Basic sound effect
rawgenai elevenlabs sfx "explosion in the distance" -o explosion.mp3

# Specific duration
rawgenai elevenlabs sfx "rain on window" -o rain.mp3 -d 10

# Looping ambient sound
rawgenai elevenlabs sfx "forest ambience with birds" -o forest.mp3 --loop -d 15

# High prompt influence for precise results
rawgenai elevenlabs sfx "single gunshot, pistol" -o gunshot.mp3 --influence 0.8
```

## API Reference

- Endpoint: `POST https://api.elevenlabs.io/v1/sound-generation`
- Auth: `xi-api-key` header
- Docs: https://elevenlabs.io/docs/api-reference/text-to-sound-effects/convert
