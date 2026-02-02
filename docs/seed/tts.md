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

With voice instruction (TTS 2.0):
```bash
rawgenai seed tts "[#用温柔的语气说] 今天天气真好" --speak
```

With emotion control:
```bash
rawgenai seed tts "[#用颤抖沙哑、带着崩溃与绝望的哭腔说] 我逆转时空九十九次救你，你却次次死于同一支暗箭。" -o emotional.mp3
```

## TTS 2.0 Features

### Voice Instructions

Control voice emotion, style, and delivery using `[#指令内容]` syntax:

- Emotion: `[#用悲伤/生气/开心的语气]`
- Style: `[#用asmr的语气]`, `[#用吵架的语气]`
- Delivery: `[#用颤抖沙哑的声音]`, `[#用低沉沧桑的语气]`

### Context Reference

Provide preceding context (not synthesized) to help the model understand emotional context:
```
[#是…是你吗？怎么看着…好像没怎么变啊？] 你头发长了…以前总说留不长，十年了…你还好吗？
```

### Voice Tags

Add action/emotion tags before sentences using `[动作/情感描述]`:
```
[怒目圆睁，冲着你大声怒吼] 放肆！我是龙族的女王！
```

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
