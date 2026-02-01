# rawgenai google stt

Speech to Text (transcription) using Google Gemini multimodal capabilities.

## Usage

```bash
rawgenai google stt <audio_file> [flags]
rawgenai google stt --file <audio_file> [flags]
```

## Examples

```bash
# Basic transcription
rawgenai google stt recording.mp3

# Specify language hint
rawgenai google stt recording.mp3 --language en

# Include timestamps
rawgenai google stt meeting.wav --timestamps

# Include speaker diarization
rawgenai google stt interview.mp3 --speakers

# Full analysis (timestamps + speakers)
rawgenai google stt podcast.mp3 --timestamps --speakers

# Output to file
rawgenai google stt recording.mp3 -o transcript.json
```

## Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `--file` | `-f` | string | - | No | Input audio file (alternative to positional arg) |
| `--output` | `-o` | string | - | No | Output file path (prints to stdout if not set) |
| `--language` | `-l` | string | - | No | Language hint (ISO 639-1 code, e.g., en, zh, ja) |
| `--timestamps` | `-t` | bool | `false` | No | Include timestamps in output |
| `--speakers` | `-s` | bool | `false` | No | Enable speaker diarization |
| `--model` | `-m` | string | `flash` | No | Model: flash |

## Supported Audio Formats

| Format | Extensions |
|--------|------------|
| MP3 | `.mp3` |
| WAV | `.wav` |
| FLAC | `.flac` |
| OGG | `.ogg` |
| AAC | `.aac`, `.m4a` |
| WebM | `.webm` |

## Output

### Basic Transcription

```json
{
  "success": true,
  "text": "Hello, this is a test recording.",
  "language": "en",
  "model": "gemini-2.5-flash"
}
```

### With Timestamps

```json
{
  "success": true,
  "text": "Hello, this is a test recording.",
  "language": "en",
  "model": "gemini-2.5-flash",
  "segments": [
    {
      "start": "00:00",
      "end": "00:02",
      "text": "Hello, this is a test recording."
    }
  ]
}
```

### With Speaker Diarization

```json
{
  "success": true,
  "text": "Hello, how are you? I'm doing great, thanks!",
  "language": "en",
  "model": "gemini-2.5-flash",
  "segments": [
    {
      "speaker": "Speaker 1",
      "start": "00:00",
      "end": "00:02",
      "text": "Hello, how are you?"
    },
    {
      "speaker": "Speaker 2",
      "start": "00:02",
      "end": "00:04",
      "text": "I'm doing great, thanks!"
    }
  ]
}
```

## Errors

```json
{
  "success": false,
  "error": {
    "code": "file_not_found",
    "message": "Audio file 'recording.mp3' does not exist"
  }
}
```

### CLI Errors (before API call)

| Code | Description |
|------|-------------|
| `missing_api_key` | GEMINI_API_KEY not set |
| `missing_input` | No audio file provided |
| `file_not_found` | Audio file does not exist |
| `unsupported_format` | Audio format not supported |
| `file_too_large` | Audio file exceeds size limit |
| `output_write_error` | Cannot write to output file |

### Gemini API Errors

| HTTP | Code | Description |
|------|------|-------------|
| 400 | `invalid_request` | Invalid request parameters |
| 400 | `invalid_audio` | Audio file is corrupted or invalid |
| 401 | `invalid_api_key` | API key is invalid or revoked |
| 403 | `permission_denied` | API key lacks required permissions |
| 429 | `rate_limit` | Too many requests |
| 429 | `quota_exceeded` | Quota exhausted |
| 500 | `server_error` | Gemini server error |
| 503 | `server_overloaded` | Gemini server overloaded |

### Network Errors

| Code | Description |
|------|-------------|
| `connection_error` | Cannot connect to Gemini API |
| `timeout` | Request timed out |

## Implementation Notes

This command uses Gemini's multimodal capabilities (not a dedicated STT API):

1. Audio file is uploaded via Files API
2. GenerateContent is called with a transcription prompt
3. Model returns structured JSON with transcription
4. Uploaded file is deleted after processing

The CLI handles prompt construction internally for a simplified user experience.
