# rawgenai openai stt

Speech to Text using OpenAI Whisper models.

## Usage

```bash
rawgenai openai stt <audio-file> [flags]
rawgenai openai stt --file <audio-file> [flags]
cat audio.mp3 | rawgenai openai stt [flags]
```

## Examples

```bash
# Basic transcription
rawgenai openai stt recording.mp3

# Specify language for better accuracy
rawgenai openai stt recording.mp3 --language en

# Use latest model
rawgenai openai stt recording.mp3 --model gpt-4o-transcribe

# With prompt to guide style/terminology
rawgenai openai stt meeting.mp3 --prompt "Technical discussion about Kubernetes and Docker"

# Get timestamps (verbose output)
rawgenai openai stt recording.mp3 --verbose

# Generate SRT subtitles
rawgenai openai stt video.mp4 --format srt -o subtitles.srt

# Generate VTT subtitles
rawgenai openai stt video.mp4 --format vtt -o subtitles.vtt

# From stdin
cat recording.mp3 | rawgenai openai stt

# Adjust temperature
rawgenai openai stt recording.mp3 --temperature 0.2
```

## Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `--file` | `-f` | string | - | No | Input audio file path |
| `--model` | `-m` | string | `whisper-1` | No | Model name |
| `--language` | `-l` | string | - | No | Language code (ISO-639-1, e.g., en, zh, ja) |
| `--prompt` | | string | - | No | Text to guide the model's style or terminology |
| `--temperature` | | float | `0` | No | Sampling temperature (0-1) |
| `--verbose` | `-v` | bool | `false` | No | Include timestamps and segments |
| `--format` | | string | `json` | No | Output format (json, text, srt, vtt) |
| `--output` | `-o` | string | - | No | Output file (required for srt/vtt formats) |

## Models

| Model | Description |
|-------|-------------|
| `whisper-1` | Standard Whisper model, supports all formats |
| `gpt-4o-transcribe` | Latest GPT-4o based, better accuracy |
| `gpt-4o-mini-transcribe` | Faster, lower cost |

## Supported Audio Formats

| Extension | Format |
|-----------|--------|
| `.mp3` | MP3 |
| `.mp4` | MP4 |
| `.mpeg` | MPEG |
| `.mpga` | MPEG Audio |
| `.m4a` | M4A |
| `.wav` | WAV |
| `.webm` | WebM |
| `.ogg` | OGG |
| `.oga` | OGA |
| `.opus` | Opus |
| `.flac` | FLAC |

**Maximum file size:** 25 MB

## Output

### Standard Output (JSON)

```json
{
  "success": true,
  "text": "Hello, this is the transcribed text from the audio file.",
  "model": "whisper-1",
  "language": "en"
}
```

### Verbose Output (--verbose)

```json
{
  "success": true,
  "text": "Hello, this is the transcribed text.",
  "model": "whisper-1",
  "language": "en",
  "duration": 5.42,
  "segments": [
    {
      "start": 0.0,
      "end": 2.5,
      "text": "Hello, this is"
    },
    {
      "start": 2.5,
      "end": 5.42,
      "text": "the transcribed text."
    }
  ]
}
```

### SRT/VTT Output (--format srt/vtt -o file)

When using `--format srt` or `--format vtt`, the subtitle content is written to the output file, and JSON response is returned:

```json
{
  "success": true,
  "file": "/path/to/subtitles.srt",
  "model": "whisper-1",
  "language": "en"
}
```

## Language Codes

Common ISO-639-1 language codes:

| Code | Language |
|------|----------|
| `en` | English |
| `zh` | Chinese |
| `ja` | Japanese |
| `ko` | Korean |
| `es` | Spanish |
| `fr` | French |
| `de` | German |
| `it` | Italian |
| `pt` | Portuguese |
| `ru` | Russian |
| `ar` | Arabic |
| `hi` | Hindi |

Specifying the language improves accuracy and reduces latency.

## Errors

```json
{
  "success": false,
  "error": {
    "code": "file_not_found",
    "message": "Audio file '/path/to/file.mp3' does not exist"
  }
}
```

### CLI Errors (before API call)

| Code | Description |
|------|-------------|
| `missing_api_key` | OPENAI_API_KEY not set |
| `missing_file` | No audio file provided |
| `file_not_found` | Audio file does not exist |
| `file_too_large` | File exceeds 25 MB limit |
| `unsupported_format` | Audio format not supported |
| `invalid_temperature` | Temperature not in range 0-1 |
| `missing_output` | --output required for srt/vtt format |
| `output_write_error` | Cannot write to output file |

### OpenAI API Errors

| HTTP | Code | Description |
|------|------|-------------|
| 400 | `invalid_model` | Model does not exist |
| 400 | `invalid_language` | Language code not supported |
| 400 | `invalid_request` | Other invalid request parameters |
| 401 | `invalid_api_key` | API key is invalid or revoked |
| 403 | `region_not_supported` | Region/country not supported |
| 413 | `file_too_large` | File exceeds size limit |
| 429 | `rate_limit` | Too many requests |
| 429 | `quota_exceeded` | Quota/credits exhausted |
| 500 | `server_error` | OpenAI server error |
| 503 | `server_overloaded` | OpenAI server overloaded |

### Network Errors

| Code | Description |
|------|-------------|
| `connection_error` | Cannot connect to OpenAI API |
| `timeout` | Request timed out |
