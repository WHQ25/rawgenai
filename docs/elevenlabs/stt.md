# ElevenLabs STT Command

## Usage

```bash
rawgenai elevenlabs stt <audio_file> [options]
```

## Input Sources

| Priority | Source | Example |
|----------|--------|---------|
| 1 | Positional argument | `rawgenai elevenlabs stt audio.mp3` |
| 2 | --file flag | `rawgenai elevenlabs stt --file audio.mp3` |
| 3 | stdin | `cat audio.mp3 \| rawgenai elevenlabs stt` |

## Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| --file | -f | string | - | No | Input audio file path |
| --model | -m | string | "scribe_v1" | No | Model: scribe_v1, scribe_v2 |
| --language | -l | string | auto | No | ISO-639 language code |
| --diarize | - | bool | false | No | Enable speaker identification |
| --speakers | - | int | - | No | Max speakers (1-32, requires --diarize) |
| --timestamps | - | string | "word" | No | Granularity: none, word, character |
| --output | -o | string | - | No | Output file (.txt, .srt, .json) |

## Models

| Model | Description |
|-------|-------------|
| scribe_v1 | Standard model |
| scribe_v2 | Improved accuracy |

## Supported Input Formats

Audio: mp3, wav, m4a, flac, ogg, webm
Video: mp4, mkv, mov, avi, webm

Max file size: 3GB

## Output Format

### Success (stdout JSON)
```json
{
  "success": true,
  "text": "Hello, this is a transcription.",
  "language": "en",
  "duration": 5.2,
  "words": [
    {"word": "Hello", "start": 0.0, "end": 0.5},
    {"word": "this", "start": 0.6, "end": 0.8}
  ]
}
```

### Success with diarization
```json
{
  "success": true,
  "text": "Hello, this is a transcription.",
  "language": "en",
  "duration": 5.2,
  "speakers": [
    {"speaker": "speaker_1", "text": "Hello", "start": 0.0, "end": 0.5},
    {"speaker": "speaker_2", "text": "this is", "start": 0.6, "end": 1.2}
  ]
}
```

### Error
```json
{
  "success": false,
  "error": {
    "code": "invalid_audio",
    "message": "Unsupported audio format"
  }
}
```

## Error Codes

### CLI Errors
| Code | Description |
|------|-------------|
| missing_input | No audio file provided |
| file_not_found | Audio file not found |
| invalid_audio | Unsupported audio format |
| file_too_large | File exceeds 3GB limit |
| invalid_model | Invalid model ID |
| invalid_parameter | Invalid flag value (timestamps, speakers) |
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

## Examples

```bash
# Basic transcription
rawgenai elevenlabs stt audio.mp3

# With language hint
rawgenai elevenlabs stt audio.mp3 -l en

# Speaker diarization
rawgenai elevenlabs stt meeting.mp3 --diarize --speakers 3

# Save to SRT file
rawgenai elevenlabs stt video.mp4 -o subtitles.srt

# Character-level timestamps
rawgenai elevenlabs stt audio.mp3 --timestamps character
```

## API Reference

- Endpoint: `POST https://api.elevenlabs.io/v1/speech-to-text`
- Auth: `xi-api-key` header
- Docs: https://elevenlabs.io/docs/api-reference/speech-to-text/convert
