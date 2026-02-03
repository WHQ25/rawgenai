# Seed TTS

Text to Speech using ByteDance Seed TTS models (V3 bidirectional streaming).

## Usage

```bash
rawgenai seed tts [text] [flags]
```

## Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--output` | `-o` | string | | Output file path |
| `--prompt-file` | | string | | Read text from file |
| `--voice` | `-V` | string | `zh_female_vv_uranus_bigtts` | Voice name |
| `--format` | | string | `mp3` | Audio format: mp3, pcm, ogg_opus |
| `--sample-rate` | | int | `24000` | Sample rate: 8000, 16000, 24000 |
| `--speed` | | int | `0` | Speech rate: -50 to 100 (0 = normal) |
| `--volume` | | int | `0` | Volume: -50 to 100 (0 = normal) |
| `--speak` | | bool | `false` | Play audio after generation |
| `--context` | | string | | Emotion/style context for TTS 2.0 |

## Environment Variables

- `SEED_APP_ID` - ByteDance API App ID (required)
- `SEED_ACCESS_TOKEN` - ByteDance API Access Token (required)

## Input Sources

Text can be provided via:
1. Positional argument: `rawgenai seed tts "Hello world"`
2. File: `rawgenai seed tts --prompt-file text.txt`
3. Stdin: `echo "Hello" | rawgenai seed tts`

## Examples

Basic usage:
```bash
rawgenai seed tts "你好世界" -o hello.mp3
```

Play without saving:
```bash
rawgenai seed tts "你好世界" --speak
```

With emotion control (TTS 2.0):
```bash
rawgenai seed tts "妈妈…妈妈她还在那里…" --context "用颤抖哭泣的语气" --speak
```

Different emotions:
```bash
# Sad tone
rawgenai seed tts "太好了！我们中奖了！" --context "用悲伤哭泣的语气" --speak

# Happy tone
rawgenai seed tts "他离开了，再也不会回来了。" --context "用开心雀跃的语气" --speak

# Storytelling tone
rawgenai seed tts "从前有一个小女孩，她住在森林边的小木屋里。" --context "用温柔讲故事的语气" --speak
```

## Emotion Control (TTS 2.0)

Use `--context` flag to control voice emotion and style. This works with TTS 2.0 voices (uranus series).

Common context examples:
- Emotion: `用悲伤的语气` (sad), `用开心雀跃的语气` (happy), `用紧张害怕的语气` (nervous)
- Style: `用温柔讲故事的语气` (storytelling), `用温暖安慰的语气` (comforting), `用asmr的语气` (ASMR)
- Speed: `说慢一点` (slower), `说快一点` (faster)
- Volume: `嗓门小一点` (quieter), `大声一点` (louder)

Note: Each `--context` applies to the entire audio. For multi-emotion stories, generate sentences separately.

## Voices

| Voice Name | voice_type | Language | Features |
|------------|------------|----------|----------|
| Vivi 2.0 | `zh_female_vv_uranus_bigtts` | 中/英 | 情感变化、指令遵循、ASMR (default) |
| 小何 2.0 | `zh_female_xiaohe_uranus_bigtts` | 中文 | 情感变化、指令遵循、ASMR |
| 云舟 2.0 | `zh_male_m191_uranus_bigtts` | 中文 | 情感变化、指令遵循、ASMR |
| 小天 2.0 | `zh_male_taocheng_uranus_bigtts` | 中文 | 情感变化、指令遵循、ASMR |

## Output Format

```json
{
  "success": true,
  "file": "/path/to/output.mp3",
  "voice": "zh_female_vv_uranus_bigtts"
}
```

## Error Codes

| Code | Description |
|------|-------------|
| `missing_text` | No text provided |
| `missing_output` | Output file required (unless --speak) |
| `missing_credentials` | SEED_APP_ID or SEED_ACCESS_TOKEN not set |
| `invalid_format` | Invalid audio format |
| `invalid_speed` | Speed out of range |
| `invalid_volume` | Volume out of range |
| `api_error` | API request failed |
| `empty_audio` | No audio data received |
