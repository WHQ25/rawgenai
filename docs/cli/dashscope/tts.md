# rawgenai dashscope tts

Text to Speech using Alibaba Qwen-TTS models via DashScope API.

Automatically selects protocol based on model name:

- Models without `-realtime` suffix → HTTP API (synchronous)
- Models with `-realtime` suffix → WebSocket API (streaming)

## Usage

```bash
rawgenai dashscope tts <text> [flags]
rawgenai dashscope tts --file <input.txt> [flags]
cat input.txt | rawgenai dashscope tts [flags]
```

## Examples

```bash
# Basic (HTTP API, WAV output)
rawgenai dashscope tts "你好，欢迎使用语音合成。" -o hello.wav

# Choose voice
rawgenai dashscope tts "Welcome to the show" --voice Serena -o welcome.wav

# From file
rawgenai dashscope tts -f script.txt -o output.wav

# From stdin
echo "Hello" | rawgenai dashscope tts -o hello.wav

# Realtime model with MP3 output
rawgenai dashscope tts "你好世界" -o hello.mp3 -m qwen3-tts-flash-realtime

# Instruct model with style control
rawgenai dashscope tts "今天天气真好" -o cheerful.mp3 \
  -m qwen3-tts-instruct-flash-realtime \
  --instructions "语速较快，充满活力，上扬语调"

# Specify language
rawgenai dashscope tts "こんにちは" -o hello.wav --voice "Ono Anna" -l Japanese

# Play audio directly
rawgenai dashscope tts "测试播放" --speak
```

## Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `--output` | `-o` | string | - | Yes* | Output file path (format from extension) |
| `--file` | `-f` | string | - | No | Input text file |
| `--voice` | - | string | `Cherry` | No | Voice name |
| `--model` | `-m` | string | `qwen3-tts-flash` | No | Model name |
| `--language` | `-l` | string | `Auto` | No | Language type |
| `--instructions` | - | string | - | No | Style instructions (instruct model only) |
| `--sample-rate` | - | int | `24000` | No | Sample rate in Hz (realtime only) |
| `--speak` | - | bool | `false` | No | Play audio after generation |

\* Required unless `--speak` is used.

## Models

### HTTP Models (synchronous)

| Model | Description |
|-------|-------------|
| `qwen3-tts-flash` | **Default.** Latest, 49 voices, 10 languages |
| `qwen3-tts-flash-2025-11-27` | Snapshot |
| `qwen3-tts-flash-2025-09-18` | Snapshot |
| `qwen-tts` | Legacy, 7 voices, Chinese/English only |

### Realtime Models (WebSocket streaming)

| Model | Description |
|-------|-------------|
| `qwen3-tts-flash-realtime` | Streaming version of flash |
| `qwen3-tts-flash-realtime-2025-11-27` | Snapshot |
| `qwen3-tts-instruct-flash-realtime` | Supports `--instructions` for style control |
| `qwen3-tts-instruct-flash-realtime-2026-01-22` | Snapshot |

## Voices

Default: `Cherry`

**All voices (49):** All support 10 languages (Chinese, English, French, German, Russian, Italian, Spanish, Portuguese, Japanese, Korean).

### General Voices

| Voice | Gender | Description |
|-------|--------|-------------|
| Cherry | F | 阳光积极、亲切自然 |
| Serena | F | 温柔小姐姐 |
| Ethan | M | 阳光温暖，标准普通话 |
| Chelsie | F | 二次元虚拟女友 |
| Momo | F | 撒娇搞怪 |
| Vivian | F | 拽拽的、可爱的小暴躁 |
| Moon | M | 率性帅气 |
| Maia | F | 知性与温柔 |
| Kai | M | 沉稳磁性 |
| Nofish | M | 不会翘舌音的设计师 |
| Bella | F | 小萝莉 |
| Jennifer | F | 品牌级美语女声 |
| Ryan | M | 节奏拉满，戏感炸裂 |
| Katerina | F | 御姐音色 |
| Aiden | M | 美语大男孩 |
| Eldric Sage | M | 沉稳睿智的老者 |
| Mia | F | 温顺乖巧 |
| Mochi | M | 聪明伶俐的小大人 |
| Bellona | F | 声音洪亮，吐字清晰 |
| Vincent | M | 沙哑烟嗓，豪情 |
| Bunny | F | 萌属性爆棚的小萝莉 |
| Neil | M | 字正腔圆的新闻主持人 |
| Elias | F | 严谨的讲师风格 |
| Arthur | M | 质朴嗓音，娓娓道来 |
| Nini | F | 软黏嗓音 |
| Ebona | F | 低语，幽暗氛围 |
| Seren | F | 温和舒缓，助眠 |
| Pip | M | 调皮捣蛋充满童真 |
| Stella | F | 甜到发腻的迷糊少女 |
| Andre | M | 声音磁性，沉稳自然 |
| Radio Gol | M | 足球解说员 |

### Foreign Language Voices

| Voice | Gender | Featured Language | Description |
|-------|--------|-------------------|-------------|
| Bodega | M | Spanish | 热情的西班牙大叔 |
| Sonrisa | F | Spanish | 热情开朗的拉美大姐 |
| Alek | M | Russian | 战斗民族的冷暖 |
| Dolce | M | Italian | 慵懒的意大利大叔 |
| Sohee | F | Korean | 温柔开朗的韩国欧尼 |
| Ono Anna | F | Japanese | 鬼灵精怪的青梅竹马 |
| Lenn | M | German | 理性的德国青年 |
| Emilien | M | French | 浪漫的法国大哥哥 |

### Dialect Voices

| Voice | Gender | Dialect | Description |
|-------|--------|---------|-------------|
| Jada | F | Shanghai | 风风火火的沪上阿姐 |
| Dylan | M | Beijing | 北京胡同少年 |
| Li | M | Nanjing | 耐心的瑜伽老师 |
| Marcus | M | Shaanxi | 老陕的味道 |
| Roy | M | Minnan | 诙谐直爽的台湾哥仔 |
| Peter | M | Tianjin | 天津相声，专业捧哏 |

> The legacy `qwen-tts` model only supports up to 7 voices. Use `qwen3-tts-flash` for all 49 voices.

## Output Formats

Format is determined by output file extension:

### HTTP Models

| Extension | Format | Description |
|-----------|--------|-------------|
| `.wav` | WAV | Only supported format (24kHz, 16-bit, mono) |

### Realtime Models

| Extension | Format | Description |
|-----------|--------|-------------|
| `.mp3` | MP3 | Compressed, general use |
| `.pcm` | PCM | Raw audio (24kHz/48kHz, 16-bit, mono) |
| `.opus` | Opus | Low latency, compressed |
| `.wav` | WAV | Uncompressed |

> When using `--speak` without `-o`, defaults to MP3 (realtime) or WAV (HTTP).

## Sample Rates (Realtime Only)

| Value | Description |
|-------|-------------|
| `24000` | 24 kHz (default) |
| `48000` | 48 kHz (higher quality) |

## Languages

| Value | Language |
|-------|----------|
| `Auto` | Auto-detect (default) |
| `Chinese` | Chinese (Mandarin) |
| `English` | English |
| `Japanese` | Japanese |
| `Korean` | Korean |
| `French` | French |
| `German` | German |
| `Russian` | Russian |
| `Italian` | Italian |
| `Spanish` | Spanish |
| `Portuguese` | Portuguese |

## Output

```json
{
  "success": true,
  "file": "/path/to/hello.wav",
  "model": "qwen3-tts-flash",
  "voice": "Cherry"
}
```

## Errors

```json
{
  "success": false,
  "error": {
    "code": "missing_text",
    "message": "no text provided, use positional argument, --file flag, or pipe from stdin"
  }
}
```

### CLI Errors (before API call)

| Code | Description |
|------|-------------|
| `missing_api_key` | DASHSCOPE_API_KEY not set |
| `missing_text` | No text provided (no argument, empty file, empty stdin) |
| `missing_output` | --output flag not provided and --speak not set |
| `file_not_found` | Input file specified by --file does not exist |
| `unsupported_format` | Output file extension not supported by model |
| `invalid_model` | Model name not recognized |
| `invalid_language` | Language type not recognized |
| `invalid_sample_rate` | Sample rate not supported (use 24000 or 48000) |
| `incompatible_instructions` | --instructions only supported by instruct models |
| `incompatible_sample_rate` | --sample-rate only supported by realtime models |
| `text_too_long` | Text exceeds 600 characters (qwen3-tts-flash) |
| `output_write_error` | Cannot write to output file |

### API Errors

| Code | Description |
|------|-------------|
| `invalid_api_key` | API key is invalid or region mismatch |
| `invalid_request` | Invalid request parameters |
| `content_policy` | Content violates safety policy |
| `rate_limit` | Too many requests |
| `server_error` | DashScope server error |

### Network Errors

| Code | Description |
|------|-------------|
| `connection_error` | Cannot connect to DashScope API |
| `timeout` | Request timed out |
| `websocket_error` | WebSocket connection failed (realtime only) |

---

## Environment Variables

| Variable | Description |
|----------|-------------|
| `DASHSCOPE_API_KEY` | DashScope API key (required) |
| `DASHSCOPE_BASE_URL` | Custom base URL (optional, default: Beijing region) |

> See [video.md](video.md) for region configuration details.

---

## Internal: Protocol Selection

The CLI determines protocol from the model name:

| Model Pattern | Protocol | Endpoint |
|---------------|----------|----------|
| `*-realtime*` | WebSocket | `wss://dashscope.aliyuncs.com/api-ws/v1/realtime?model={model}` |
| Others | HTTP | `POST {base_url}/services/aigc/multimodal-generation/generation` |

### HTTP Flow

1. POST with text, voice, language → response with `audio.url`
2. Download audio URL → save to output file
3. Audio URL valid for 24 hours

### WebSocket Flow

1. Connect to `wss://...?model={model}` with auth header
2. Send `session.update` (voice, format, sample_rate, language, instructions)
3. Send `input_text_buffer.append` with text
4. Send `input_text_buffer.commit` (commit mode)
5. Receive `response.audio.delta` events (base64 audio chunks)
6. Write decoded audio to output file
7. Send `session.finish` on `response.done`

WebSocket mode uses `commit` (client-controlled) for CLI, not `server_commit`.
