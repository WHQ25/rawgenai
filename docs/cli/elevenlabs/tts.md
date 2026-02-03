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
| --output | -o | string | - | Yes | Output file path (.mp3, .wav, .pcm, .opus) |
| --voice | -v | string | "Rachel" | No | Voice name or ID |
| --model | -m | string | "eleven_multilingual_v2" | No | Model ID |
| --format | -f | string | "mp3_44100_128" | No | Output format |
| --language | -l | string | - | No | Language code (ISO 639-1, e.g., "en", "zh") |
| --stability | - | float | 0.5 | No | Voice stability (0.0-1.0) |
| --similarity | - | float | 0.75 | No | Similarity boost (0.0-1.0) |
| --style | - | float | 0.0 | No | Style exaggeration (0.0-1.0) |
| --speed | - | float | 1.0 | No | Speaking speed (0.25-4.0) |
| --speaker-boost | - | bool | true | No | Boost similarity to original voice |
| --text-normalization | - | string | "auto" | No | Text normalization: auto, on, off |
| --stream | - | bool | false | No | Use streaming mode for lower latency |
| --file | - | string | - | No | Input text file |
| --speak | - | bool | false | No | Play audio after generation |

## Streaming Mode

Use `--stream` flag to enable streaming mode, which provides lower latency by returning audio chunks as they're generated instead of waiting for the full audio to be synthesized.

When combined with `--speak`, audio plays directly from the stream without waiting for the full download:

```bash
# Stream and play immediately (lowest latency)
rawgenai elevenlabs tts "Hello world" --stream --speak

# Stream, play, and save to file simultaneously
rawgenai elevenlabs tts "Hello world" --stream --speak -o hello.mp3
```

## Models

| Model ID | Description | Languages | Character Limit |
|----------|-------------|-----------|-----------------|
| eleven_v3 | Most expressive, emotionally rich | 70+ | 5,000 |
| eleven_multilingual_v2 | Lifelike, consistent quality | 29 | 10,000 |
| eleven_flash_v2_5 | Ultra-low latency (~75ms) | 32 | 40,000 |
| eleven_turbo_v2_5 | Balanced quality and speed | 32 | 40,000 |

## Output Formats

### MP3
| Format | Description |
|--------|-------------|
| mp3_22050_32 | MP3, 22kHz, 32kbps |
| mp3_24000_48 | MP3, 24kHz, 48kbps |
| mp3_44100_32 | MP3, 44.1kHz, 32kbps |
| mp3_44100_64 | MP3, 44.1kHz, 64kbps |
| mp3_44100_96 | MP3, 44.1kHz, 96kbps |
| mp3_44100_128 | MP3, 44.1kHz, 128kbps (default) |
| mp3_44100_192 | MP3, 44.1kHz, 192kbps (Creator+) |

### Opus
| Format | Description |
|--------|-------------|
| opus_48000_32 | Opus, 48kHz, 32kbps |
| opus_48000_64 | Opus, 48kHz, 64kbps |
| opus_48000_96 | Opus, 48kHz, 96kbps |
| opus_48000_128 | Opus, 48kHz, 128kbps |
| opus_48000_192 | Opus, 48kHz, 192kbps |

### PCM (raw audio)
| Format | Description |
|--------|-------------|
| pcm_8000 | PCM, 8kHz |
| pcm_16000 | PCM, 16kHz |
| pcm_22050 | PCM, 22.05kHz |
| pcm_24000 | PCM, 24kHz |
| pcm_32000 | PCM, 32kHz |
| pcm_44100 | PCM, 44.1kHz (Pro+) |
| pcm_48000 | PCM, 48kHz |

### WAV
| Format | Description |
|--------|-------------|
| wav_8000 | WAV, 8kHz |
| wav_16000 | WAV, 16kHz |
| wav_22050 | WAV, 22.05kHz |
| wav_24000 | WAV, 24kHz |
| wav_32000 | WAV, 32kHz |
| wav_44100 | WAV, 44.1kHz (Pro+) |
| wav_48000 | WAV, 48kHz |

### Telephony
| Format | Description |
|--------|-------------|
| alaw_8000 | A-law, 8kHz (telephony) |
| ulaw_8000 | μ-law, 8kHz (Twilio) |

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
  "characters": 42,
  "stream": false
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
| invalid_text_normalization | Invalid text normalization value (must be auto, on, off) |
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

# Use the new v3 model (most expressive)
rawgenai elevenlabs tts "An emotional story" -o story.mp3 -m eleven_v3

# With voice selection
rawgenai elevenlabs tts "Welcome" -o welcome.mp3 -v "Josh"

# High quality settings
rawgenai elevenlabs tts "Premium audio" -o premium.mp3 -f mp3_44100_192

# Opus format (good for streaming)
rawgenai elevenlabs tts "Stream this" -o stream.opus -f opus_48000_128

# WAV format (lossless)
rawgenai elevenlabs tts "High fidelity" -o hifi.wav -f wav_48000

# Specify language for better pronunciation
rawgenai elevenlabs tts "你好世界" -o chinese.mp3 -l zh

# Adjust voice parameters
rawgenai elevenlabs tts "Calm narration" -o calm.mp3 --stability 0.8 --speed 0.9

# Disable text normalization (read numbers literally)
rawgenai elevenlabs tts "Call 1-800-555-0123" -o phone.mp3 --text-normalization off

# Streaming mode for lower latency
rawgenai elevenlabs tts "Quick response" -o quick.mp3 --stream

# Stream and play immediately
rawgenai elevenlabs tts "Play this now" --stream --speak
```

## API Reference

- Convert endpoint: `POST https://api.elevenlabs.io/v1/text-to-speech/{voice_id}`
- Stream endpoint: `POST https://api.elevenlabs.io/v1/text-to-speech/{voice_id}/stream`
- Auth: `xi-api-key` header
- Docs: https://elevenlabs.io/docs/api-reference/text-to-speech/convert
