# rawgenai google tts

Text to Speech using Google Gemini TTS models.

> **Note:** Unlike other TTS providers (OpenAI, ElevenLabs), Gemini TTS accepts a **prompt** that can include style instructions to control voice characteristics like tone, pace, and emotion.

> **Important:** For complex sentences, use explicit style prefixes like `Say:` or `Read:` to ensure proper TTS behavior. Without these, the model may misinterpret certain phrases as instructions rather than text to speak.

## Usage

```bash
rawgenai google tts <prompt> [flags]
rawgenai google tts --prompt-file <input.txt> [flags]
cat input.txt | rawgenai google tts [flags]
```

## Examples

### Basic Usage

```bash
# Simple text (plain text works too)
rawgenai google tts "Hello, world!" -o hello.wav

# With voice selection
rawgenai google tts "Welcome to the show" --voice Kore -o welcome.wav
```

### Prompt-based Style Control

Gemini TTS supports natural language prompts to control speech style:

```bash
# Cheerful tone
rawgenai google tts "Say cheerfully: Hello everyone, welcome to the show!" -o cheerful.wav

# Whispering
rawgenai google tts "Whisper: This is a secret message" -o whisper.wav

# Slow and dramatic
rawgenai google tts "Say slowly and dramatically: The end is near..." -o dramatic.wav

# News anchor style
rawgenai google tts "Read like a news anchor: Breaking news today..." -o news.wav

# Excited
rawgenai google tts "Say excitedly: We won the championship!" -o excited.wav
```

### Advanced Examples

```bash
# Using Pro model for higher quality
rawgenai google tts "Important announcement" --model pro --voice Puck -o announcement.wav

# Multi-speaker conversation
rawgenai google tts "Joe: Hi there!\nJane: Hello!" --speakers "Joe=Kore,Jane=Puck" -o conversation.wav

# From file
rawgenai google tts --prompt-file script.txt -o output.wav

# From stdin
echo "Say calmly: Hello" | rawgenai google tts -o hello.wav

# Play directly without saving
rawgenai google tts "Hello world" --speak

# Convert to MP3 (requires ffmpeg)
rawgenai google tts "Hello" -o hello.wav && ffmpeg -i hello.wav hello.mp3
```

## Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `--output` | `-o` | string | - | No* | Output file path (.wav) |
| `--prompt-file` | - | string | - | No | Input prompt file |
| `--voice` | `-v` | string | `Kore` | No | Voice name (single speaker) |
| `--speakers` | - | string | - | No | Multi-speaker config: "Name1=Voice1,Name2=Voice2" |
| `--model` | `-m` | string | `flash` | No | Model: flash, pro |
| `--speak` | - | bool | `false` | No | Play audio after generation |

*Required unless `--speak` is used.

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

## Troubleshooting

### "Model tried to generate text" Error

If you see this error:
```json
{"success":false,"error":{"code":"api_error","message":"Model tried to generate text..."}}
```

The Gemini TTS model misinterpreted your text as an instruction. Fix by adding an explicit style prefix:

```bash
# ❌ May fail
rawgenai google tts "This is a test of the speak feature." -o test.wav

# ✅ Works
rawgenai google tts "Say: This is a test of the speak feature." -o test.wav
rawgenai google tts "Read aloud: This is a test of the speak feature." -o test.wav
```

Common prefixes that work:
- `Say:` - neutral reading
- `Read aloud:` - clear narration
- `Speak:` - natural speech
- `Announce:` - formal tone

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
| `file_not_found` | Input file specified by --prompt-file does not exist |
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
