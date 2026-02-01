# rawgenai google tts

Text to Speech using Google Gemini TTS models.

## Usage

```bash
rawgenai google tts <text> [flags]
rawgenai google tts --file <input.txt> [flags]
cat input.txt | rawgenai google tts [flags]
```

## Examples

```bash
# Basic
rawgenai google tts "Hello, world!" -o hello.wav

# With voice selection
rawgenai google tts "Welcome to the show" --voice Kore -o welcome.wav

# Using Pro model for higher quality
rawgenai google tts "Important announcement" --model pro --voice Puck -o announcement.wav

# Multi-speaker conversation
rawgenai google tts "Joe: Hi there!\nJane: Hello!" --speakers "Joe=Kore,Jane=Puck" -o conversation.wav

# From file
rawgenai google tts --file script.txt -o output.wav

# From stdin
echo "Hello" | rawgenai google tts -o hello.wav

# Convert to MP3 (requires ffmpeg)
rawgenai google tts "Hello" -o hello.wav && ffmpeg -i hello.wav hello.mp3
```

## Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `--output` | `-o` | string | - | Yes | Output file path (.wav) |
| `--file` | - | string | - | No | Input text file |
| `--voice` | `-v` | string | `Kore` | No | Voice name (single speaker) |
| `--speakers` | - | string | - | No | Multi-speaker config: "Name1=Voice1,Name2=Voice2" |
| `--model` | `-m` | string | `flash` | No | Model: flash, pro |

## Models

| Model | Model ID | Description |
|-------|----------|-------------|
| `flash` | `gemini-2.5-flash-preview-tts` | Fast, efficient (default) |
| `pro` | `gemini-2.5-pro-preview-tts` | Higher quality |

## Voices

30 prebuilt voices available:

| Voice | Characteristic |
|-------|----------------|
| `Zephyr` | Bright |
| `Puck` | Upbeat |
| `Charon` | Informative |
| `Kore` | Firm (default) |
| `Fenrir` | Excitable |
| `Leda` | Youthful |
| `Orus` | Firm |
| `Aoede` | Breezy |
| `Callirrhoe` | Easy-going |
| `Autonoe` | Bright |
| `Enceladus` | Breathy |
| `Iapetus` | Clear |
| `Umbriel` | Easy-going |
| `Algieba` | Smooth |
| `Despina` | Smooth |
| `Erinome` | Clear |
| `Algenib` | Gravelly |
| `Rasalgethi` | Informative |
| `Laomedeia` | Upbeat |
| `Achernar` | Soft |
| `Alnilam` | Firm |
| `Schedar` | Even |
| `Gacrux` | Mature |
| `Pulcherrima` | Forward |
| `Achird` | Friendly |
| `Zubenelgenubi` | Casual |
| `Vindemiatrix` | Gentle |
| `Sadachbia` | Lively |
| `Sadaltager` | Knowledgeable |
| `Sulafat` | Warm |

## Multi-Speaker Mode

For conversations with multiple speakers, use `--speakers` flag:

```bash
# Format: "SpeakerName1=VoiceName1,SpeakerName2=VoiceName2"
rawgenai google tts "Joe: How's it going?\nJane: Not bad!" --speakers "Joe=Kore,Jane=Puck" -o chat.wav
```

**Requirements:**
- Speaker names in text must match names in `--speakers` config
- Maximum 2 speakers supported
- Cannot use `--voice` and `--speakers` together

## Output Format

Output is always WAV format (PCM 24kHz, 16-bit, mono).

**Note:** Gemini TTS outputs raw PCM audio. The CLI automatically converts to WAV format. For other formats (MP3, etc.), use ffmpeg post-processing.

## Output

```json
{
  "success": true,
  "file": "/path/to/hello.wav",
  "model": "gemini-2.5-flash-preview-tts",
  "voice": "Kore"
}
```

Multi-speaker output:

```json
{
  "success": true,
  "file": "/path/to/conversation.wav",
  "model": "gemini-2.5-flash-preview-tts",
  "speakers": {
    "Joe": "Kore",
    "Jane": "Puck"
  }
}
```

## Errors

```json
{
  "success": false,
  "error": {
    "code": "invalid_voice",
    "message": "Voice 'xxx' is not a valid prebuilt voice"
  }
}
```

### CLI Errors (before API call)

| Code | Description |
|------|-------------|
| `missing_api_key` | GEMINI_API_KEY not set |
| `missing_text` | No text provided (no argument, empty file, empty stdin) |
| `missing_output` | --output flag not provided |
| `file_not_found` | Input file specified by --file does not exist |
| `unsupported_format` | Output file extension not .wav |
| `invalid_voice` | Voice name not in prebuilt voices list |
| `invalid_model` | Model not flash or pro |
| `invalid_speakers` | --speakers format invalid or speaker not found in text |
| `conflicting_flags` | Cannot use --voice and --speakers together |
| `too_many_speakers` | More than 2 speakers specified |
| `output_write_error` | Cannot write to output file |

### Gemini API Errors

| HTTP | Code | Description |
|------|------|-------------|
| 400 | `invalid_request` | Invalid request parameters |
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
