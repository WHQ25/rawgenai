# rawgenai openai tts

Text to Speech using OpenAI TTS models.

## Usage

```bash
rawgenai openai tts <text> [flags]
rawgenai openai tts --file <input.txt> [flags]
cat input.txt | rawgenai openai tts [flags]
```

## Examples

```bash
# Basic
rawgenai openai tts "Hello, world!" -o hello.mp3

# With voice and instructions
rawgenai openai tts "Welcome to the show" --voice coral --instructions "Speak cheerfully and energetically" -o welcome.mp3

# High quality, slower
rawgenai openai tts "Important announcement" --model tts-1-hd --voice onyx -o announcement.mp3

# From file
rawgenai openai tts --file script.txt -o output.mp3

# From stdin
echo "Hello" | rawgenai openai tts -o hello.mp3

# Different format
rawgenai openai tts "Hello" -o hello.flac

# Adjust speed
rawgenai openai tts "Breaking news" --speed 1.2 -o news.mp3
```

## Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `--output` | `-o` | string | - | Yes | Output file path (format from extension) |
| `--file` | - | string | - | No | Input text file |
| `--voice` | - | string | `alloy` | No | Voice name or custom voice ID |
| `--model` | `-m` | string | `gpt-4o-mini-tts` | No | Model name |
| `--instructions` | - | string | - | No | Voice style instructions (gpt-4o-mini-tts only) |
| `--speed` | - | float | `1.0` | No | Speed (0.25 - 4.0) |

## Models

| Model | Description |
|-------|-------------|
| `gpt-4o-mini-tts` | Latest, supports instructions for style control |
| `tts-1` | Lower latency, standard quality |
| `tts-1-hd` | Higher quality, slower |

## Voices

**Recommended:** `marin`, `cedar`

**All voices (13):**

| Voice | Available in |
|-------|--------------|
| `alloy` | all models |
| `ash` | all models |
| `ballad` | gpt-4o-mini-tts only |
| `coral` | all models |
| `echo` | all models |
| `fable` | all models |
| `nova` | all models |
| `onyx` | all models |
| `sage` | all models |
| `shimmer` | all models |
| `verse` | gpt-4o-mini-tts only |
| `marin` | gpt-4o-mini-tts only |
| `cedar` | gpt-4o-mini-tts only |

**Custom voice:** Pass a voice ID (e.g., `voice_123abc`) to use a custom voice created via OpenAI API.

## Output Formats

Format is determined by output file extension:

| Extension | Format | Description |
|-----------|--------|-------------|
| `.mp3` | MP3 | Default, general use |
| `.opus` | Opus | Low latency, streaming |
| `.aac` | AAC | Mobile preferred |
| `.flac` | FLAC | Lossless |
| `.wav` | WAV | Uncompressed, low latency |
| `.pcm` | PCM | Raw 24kHz 16-bit |

## Output

```json
{
  "success": true,
  "file": "/path/to/hello.mp3",
  "model": "gpt-4o-mini-tts",
  "voice": "coral"
}
```

## Errors

```json
{
  "success": false,
  "error": {
    "code": "invalid_voice",
    "message": "Voice 'xxx' is not available for model 'tts-1'"
  }
}
```

### CLI Errors (before API call)

| Code | Description |
|------|-------------|
| `missing_api_key` | OPENAI_API_KEY not set |
| `missing_text` | No text provided (no argument, empty file, empty stdin) |
| `missing_output` | --output flag not provided |
| `file_not_found` | Input file specified by --file does not exist |
| `unsupported_format` | Output file extension not supported |
| `invalid_speed` | Speed value not in range 0.25-4.0 |
| `output_write_error` | Cannot write to output file |

### OpenAI API Errors

| HTTP | Code | Description |
|------|------|-------------|
| 400 | `invalid_voice` | Voice does not exist or not supported by model |
| 400 | `invalid_model` | Model does not exist |
| 400 | `text_too_long` | Text exceeds 4096 characters |
| 400 | `invalid_request` | Other invalid request parameters |
| 401 | `invalid_api_key` | API key is invalid or revoked |
| 403 | `region_not_supported` | Region/country not supported |
| 429 | `rate_limit` | Too many requests |
| 429 | `quota_exceeded` | Quota/credits exhausted |
| 500 | `server_error` | OpenAI server error |
| 503 | `server_overloaded` | OpenAI server overloaded |

### Network Errors

| Code | Description |
|------|-------------|
| `connection_error` | Cannot connect to OpenAI API |
| `timeout` | Request timed out |
