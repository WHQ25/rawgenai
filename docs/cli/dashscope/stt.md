# rawgenai dashscope stt

Speech to Text using Alibaba DashScope ASR models.

Supports three recognition modes:

- **Synchronous** — `qwen3-asr-flash` via HTTP API, short audio (≤5min)
- **Realtime** — `paraformer-realtime-v2`, `fun-asr-realtime` etc. via WebSocket, streaming from local file
- **Async file transcription** — `paraformer-v2`, `fun-asr`, `qwen3-asr-flash-filetrans` via async REST API, long audio URLs (≤12h)

Protocol is auto-selected based on model name:

- Models with `-realtime` suffix → WebSocket streaming
- `qwen3-asr-flash` → HTTP sync API
- Others → async REST (only via `create` subcommand)

## Commands

| Command | Description |
|---------|-------------|
| `stt` (default) | Recognize local audio file (sync or realtime models) |
| `stt create` | Submit async transcription task for audio URL(s) |
| `stt status` | Query async task status and get results |

---

## dashscope stt

Recognize a local audio file. Auto-selects sync HTTP or WebSocket based on model.

### Usage

```bash
rawgenai dashscope stt <audio_file> [flags]
rawgenai dashscope stt --file <audio_file> [flags]
cat audio.wav | rawgenai dashscope stt [flags]
```

### Examples

```bash
# Basic transcription (qwen3-asr-flash, sync HTTP)
rawgenai dashscope stt recording.mp3

# Specify language for better accuracy
rawgenai dashscope stt recording.mp3 -l zh

# Disable ITN (inverse text normalization)
rawgenai dashscope stt recording.mp3 --no-itn

# Save to file
rawgenai dashscope stt recording.mp3 -o transcript.json

# Realtime model with word-level timestamps
rawgenai dashscope stt meeting.wav -m paraformer-realtime-v2 --verbose

# Fun-ASR realtime
rawgenai dashscope stt recording.wav -m fun-asr-realtime

# Realtime with language hints
rawgenai dashscope stt recording.wav -m paraformer-realtime-v2 --language-hints zh,en

# Qwen-ASR realtime (30+ languages, emotion detection)
rawgenai dashscope stt recording.wav -m qwen3-asr-flash-realtime

# Enable hot words
rawgenai dashscope stt meeting.wav -m paraformer-realtime-v2 --vocabulary-id vocab_xxx

# Enable disfluency removal
rawgenai dashscope stt meeting.wav -m paraformer-realtime-v2 --disfluency-removal

# From stdin
cat recording.wav | rawgenai dashscope stt

# From file flag
rawgenai dashscope stt -f recording.mp3
```

### Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `--file` | `-f` | string | - | No | Input audio file path |
| `--model` | `-m` | string | `qwen3-asr-flash` | No | Model name |
| `--language` | `-l` | string | - | No | Language code (zh, en, ja, etc.) |
| `--no-itn` | - | bool | `false` | No | Disable inverse text normalization |
| `--verbose` | `-v` | bool | `false` | No | Include timestamps and segments |
| `--output` | `-o` | string | - | No | Output file path |
| `--vocabulary-id` | - | string | - | No | Hot words vocabulary ID (realtime only) |
| `--disfluency-removal` | - | bool | `false` | No | Remove filler words (paraformer/fun-asr realtime only) |
| `--language-hints` | - | string | - | No | Comma-separated language hints (paraformer-realtime-v2 only) |
| `--sample-rate` | - | int | auto | No | Sample rate in Hz (realtime only, auto-detected from file) |

### Sync Models (HTTP API)

| Model | Max Duration | Max Size | Languages | Description |
|-------|-------------|---------|-----------|-------------|
| `qwen3-asr-flash` | 5 min | 10 MB | 30+ | **Default.** Sync, emotion detection |
| `qwen3-asr-flash-2025-09-08` | 5 min | 10 MB | 30+ | Snapshot |

### Realtime Models (WebSocket)

| Model | Sample Rate | Languages | Protocol | Description |
|-------|------------|-----------|----------|-------------|
| `paraformer-realtime-v2` | any | zh, en, ja, yue, ko, de, fr, ru | run-task | Multilingual, timestamps, hot words |
| `paraformer-realtime-v1` | 16kHz | zh | run-task | Chinese only |
| `paraformer-realtime-8k-v2` | 8kHz | zh | run-task | Emotion recognition, phone calls |
| `paraformer-realtime-8k-v1` | 8kHz | zh | run-task | Basic 8k |
| `fun-asr-realtime` | 16kHz | zh, en, ja | run-task | Stable (= fun-asr-realtime-2025-11-07) |
| `fun-asr-realtime-2025-11-07` | 16kHz | zh, en, ja | run-task | VAD improvements |
| `fun-asr-realtime-2025-09-15` | 16kHz | zh, en, ja | run-task | Snapshot |
| `qwen3-asr-flash-realtime` | 8k/16kHz | 30+ | session.update | Emotion detection, no timestamps |
| `qwen3-asr-flash-realtime-2025-10-27` | 8k/16kHz | 30+ | session.update | Snapshot |

### Supported Audio Formats

#### Sync Models (qwen3-asr-flash)

All common audio formats. Max 10 MB, max 5 minutes.

Input methods: local file (base64-encoded), URL, stdin.

#### Realtime Models

| Format | Extension |
|--------|-----------|
| PCM | `.pcm` |
| WAV | `.wav` |
| MP3 | `.mp3` |
| Opus | `.opus` |
| Speex | `.spx` |
| AAC | `.aac` |
| AMR | `.amr` |

Audio must be **mono**. When sending via WebSocket, the CLI reads the file and sends ~100ms chunks.

### Output

#### Standard Output

```json
{
  "success": true,
  "text": "你好，这是一段测试录音。",
  "model": "qwen3-asr-flash",
  "language": "zh"
}
```

#### With Emotion (qwen3-asr-flash, qwen3-asr-flash-realtime, paraformer-realtime-8k-v2)

```json
{
  "success": true,
  "text": "你好，这是一段测试录音。",
  "model": "qwen3-asr-flash",
  "language": "zh",
  "emotion": "neutral"
}
```

Supported emotions: `neutral`, `happy`, `sad`, `angry`, `surprise`, `calm`, `fearful`, `disgusted`

#### Verbose Output (--verbose, realtime models only)

```json
{
  "success": true,
  "text": "你好，这是一段测试录音。",
  "model": "paraformer-realtime-v2",
  "duration": 3.5,
  "segments": [
    {
      "start": 0.17,
      "end": 3.50,
      "text": "你好，这是一段测试录音。",
      "words": [
        {"text": "你好", "start": 0.17, "end": 0.45, "punctuation": "，"},
        {"text": "这是", "start": 0.50, "end": 0.80, "punctuation": ""},
        {"text": "一段", "start": 0.85, "end": 1.10, "punctuation": ""},
        {"text": "测试", "start": 1.15, "end": 1.50, "punctuation": ""},
        {"text": "录音", "start": 1.55, "end": 1.90, "punctuation": "。"}
      ]
    }
  ]
}
```

> `--verbose` with `qwen3-asr-flash` (sync) has no effect since the sync API does not return timestamps. `--verbose` with `qwen3-asr-flash-realtime` also has no effect (no timestamp support).

### Flag Compatibility

| Flag | qwen3-asr-flash | paraformer-realtime-* | fun-asr-realtime | qwen3-asr-flash-realtime |
|------|----------------|----------------------|-----------------|-------------------------|
| `--language` | ✅ | ❌ (use --language-hints) | ❌ (use --language-hints) | ✅ |
| `--no-itn` | ✅ | ❌ | ❌ | ❌ |
| `--verbose` | ignored | ✅ | ✅ | ignored |
| `--vocabulary-id` | ❌ | ✅ | ✅ | ❌ |
| `--disfluency-removal` | ❌ | ✅ (paraformer only) | ✅ | ❌ |
| `--language-hints` | ❌ | ✅ (v2 only) | ✅ | ❌ |
| `--sample-rate` | ❌ | ✅ | ✅ | ✅ |

---

## dashscope stt create

Submit async transcription task(s) for audio file URL(s). For long audio files (up to 12 hours, 2 GB).

### Usage

```bash
rawgenai dashscope stt create <url> [url...] [flags]
```

### Examples

```bash
# Single file with default model (paraformer-v2)
rawgenai dashscope stt create https://example.com/meeting.wav

# Multiple files
rawgenai dashscope stt create https://example.com/part1.wav https://example.com/part2.wav

# Specify model
rawgenai dashscope stt create https://example.com/audio.mp3 -m fun-asr

# Multilingual model
rawgenai dashscope stt create https://example.com/audio.mp3 -m fun-asr-mtl

# Qwen-ASR file transcription
rawgenai dashscope stt create https://example.com/audio.mp3 -m qwen3-asr-flash-filetrans

# With language hints
rawgenai dashscope stt create https://example.com/audio.mp3 --language-hints zh,en

# With hot words
rawgenai dashscope stt create https://example.com/audio.mp3 --vocabulary-id vocab_xxx

# Enable speaker diarization
rawgenai dashscope stt create https://example.com/meeting.wav --diarize --speakers 3

# Enable disfluency removal
rawgenai dashscope stt create https://example.com/audio.mp3 --disfluency-removal
```

### Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `--model` | `-m` | string | `paraformer-v2` | No | Model name |
| `--language-hints` | - | string | - | No | Comma-separated language hints (paraformer-v2 only) |
| `--vocabulary-id` | - | string | - | No | Hot words vocabulary ID |
| `--disfluency-removal` | - | bool | `false` | No | Remove filler words (paraformer/fun-asr) |
| `--diarize` | - | bool | `false` | No | Enable speaker diarization |
| `--speakers` | - | int | - | No | Reference speaker count, 2-100 (requires --diarize) |
| `--channel` | - | int[] | `[0]` | No | Audio channel indices |
| `--itn` | - | bool | `false` | No | Enable ITN (qwen3-asr-flash-filetrans only) |
| `--words` | - | bool | `false` | No | Enable word-level timestamps (qwen3-asr-flash-filetrans only) |

### Async Models

| Model | Languages | Description |
|-------|-----------|-------------|
| `paraformer-v2` | **Default.** zh, yue, wu, min, en, ja, ko, de, fr, ru | Multilingual, language_hints, hot words |
| `paraformer-v1` | zh, en | Chinese/English |
| `paraformer-8k-v2` | zh | 8kHz, phone calls |
| `paraformer-8k-v1` | zh | Legacy 8k |
| `paraformer-mtl-v1` | 10+ languages | Multilingual (≥16kHz) |
| `fun-asr` | zh (普通话+方言), en, ja | Stable (= fun-asr-2025-11-07), song recognition |
| `fun-asr-mtl` | zh, yue, en, ja, th, vi, id | Multilingual |
| `qwen3-asr-flash-filetrans` | 30+ languages | Up to 12h, emotion detection |

### Audio Requirements

- **Input**: public URL only (HTTP/HTTPS), no local file upload or base64
- **Formats**: aac, amr, avi, flac, flv, m4a, mkv, mov, mp3, mp4, mpeg, ogg, opus, wav, webm, wma, wmv
- **Max file size**: 2 GB
- **Max duration**: 12 hours
- **Max URLs per request**: 100

### Output

```json
{
  "success": true,
  "task_id": "f86ec806-4d73-485f-a24f-xxxxxxxxxxxx",
  "status": "pending"
}
```

### Flag Compatibility

| Flag | paraformer-v2 | paraformer-v1 | paraformer-mtl-v1 | fun-asr | qwen3-asr-flash-filetrans |
|------|--------------|--------------|-------------------|---------|--------------------------|
| `--language-hints` | ✅ | ❌ | ❌ | ❌ | ❌ |
| `--vocabulary-id` | ✅ | ✅ | ✅ | ✅ | ❌ |
| `--disfluency-removal` | ✅ | ✅ | ✅ | ❌ | ❌ |
| `--diarize` | ✅ | ✅ | ✅ | ✅ | ❌ |
| `--itn` | ❌ | ❌ | ❌ | ❌ | ✅ |
| `--words` | ❌ | ❌ | ❌ | ❌ | ✅ |

---

## dashscope stt status

Query the status of an async transcription task. Returns transcription text when task succeeds.

### Usage

```bash
rawgenai dashscope stt status <task_id>
rawgenai dashscope stt status <task_id> -v    # Show full details
rawgenai dashscope stt status <task_id> -o transcript.json  # Save results to file
```

### Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `--verbose` | `-v` | bool | `false` | No | Show full output including URLs and per-file details |
| `--output` | `-o` | string | - | No | Save transcription result(s) to file |

### Output

**Pending/Running:**
```json
{
  "success": true,
  "task_id": "xxx",
  "status": "running"
}
```

**Succeeded (single file):**
```json
{
  "success": true,
  "task_id": "xxx",
  "status": "succeeded",
  "text": "Hello world, 这里是阿里巴巴语音实验室。",
  "duration": 5
}
```

**Succeeded (multiple files):**
```json
{
  "success": true,
  "task_id": "xxx",
  "status": "succeeded",
  "results": [
    {
      "file_url": "https://xxx1.wav",
      "text": "Hello world.",
      "status": "succeeded"
    },
    {
      "file_url": "https://xxx2.wav",
      "text": "你好世界。",
      "status": "succeeded"
    }
  ],
  "duration": 9
}
```

**Succeeded with --verbose:**
```json
{
  "success": true,
  "task_id": "xxx",
  "status": "succeeded",
  "results": [
    {
      "file_url": "https://xxx1.wav",
      "transcription_url": "https://xxx1.json",
      "status": "succeeded",
      "text": "Hello world, 这里是阿里巴巴语音实验室。",
      "segments": [
        {
          "channel_id": 0,
          "text": "Hello world, 这里是阿里巴巴语音实验室。",
          "sentences": [
            {
              "start": 170,
              "end": 4950,
              "text": "Hello world, 这里是阿里巴巴语音实验室。"
            }
          ]
        }
      ]
    }
  ],
  "duration": 9,
  "metrics": {
    "total": 1,
    "succeeded": 1,
    "failed": 0
  }
}
```

**With -o flag:** saves full transcription result to file, JSON output includes `file` field:
```json
{
  "success": true,
  "task_id": "xxx",
  "status": "succeeded",
  "file": "/absolute/path/transcript.json",
  "duration": 9
}
```

**Partial failure (some files failed):**
```json
{
  "success": true,
  "task_id": "xxx",
  "status": "succeeded",
  "results": [
    {
      "file_url": "https://xxx1.wav",
      "text": "Hello world.",
      "status": "succeeded"
    },
    {
      "file_url": "https://xxx2.wav",
      "status": "failed",
      "error": "The audio file cannot be downloaded."
    }
  ],
  "duration": 5
}
```

### Status Values

| Status | API Status | Description |
|--------|-----------|-------------|
| `pending` | PENDING | Task is waiting in queue |
| `running` | RUNNING | Transcription in progress |
| `succeeded` | SUCCEEDED | Transcription completed |
| `failed` | FAILED | Transcription failed |

### Notes

- Transcription result URLs are valid for **24 hours**
- When a task has multiple files, overall status is `succeeded` if any file succeeds. Check individual `status` fields for per-file results.

---

## Language Codes

### qwen3-asr-flash / qwen3-asr-flash-realtime (--language)

| Code | Language |
|------|----------|
| `zh` | Chinese (Mandarin) |
| `en` | English |
| `ja` | Japanese |
| `ko` | Korean |
| `fr` | French |
| `de` | German |
| `ru` | Russian |
| `es` | Spanish |
| `it` | Italian |
| `pt` | Portuguese |
| `ar` | Arabic |
| `hi` | Hindi |
| `id` | Indonesian |
| `th` | Thai |
| `tr` | Turkish |
| `vi` | Vietnamese |
| ... | 30+ languages total |

### paraformer-realtime-v2 / paraformer-v2 (--language-hints)

| Code | Language |
|------|----------|
| `zh` | Chinese |
| `en` | English |
| `ja` | Japanese |
| `yue` | Cantonese |
| `ko` | Korean |
| `de` | German |
| `fr` | French |
| `ru` | Russian |

### fun-asr-realtime (--language-hints)

| Code | Language |
|------|----------|
| `zh` | Chinese |
| `en` | English |
| `ja` | Japanese |

---

## Errors

```json
{
  "success": false,
  "error": {
    "code": "file_not_found",
    "message": "audio file 'recording.mp3' does not exist"
  }
}
```

### CLI Errors (before API call)

| Code | Description |
|------|-------------|
| `missing_api_key` | DASHSCOPE_API_KEY not set |
| `missing_input` | No audio file provided (no argument, no --file, no stdin) |
| `missing_url` | No audio URL provided (create subcommand) |
| `missing_task_id` | Task ID not provided (status subcommand) |
| `file_not_found` | Audio file does not exist |
| `file_too_large` | Audio file exceeds size limit (10 MB for sync) |
| `invalid_model` | Model name not recognized |
| `invalid_language` | Language code not supported |
| `invalid_language_hints` | Language hints format invalid or unsupported code |
| `invalid_speakers` | Speaker count out of range (2-100) |
| `incompatible_language` | --language not supported by this model (use --language-hints) |
| `incompatible_language_hints` | --language-hints not supported by this model (use --language) |
| `incompatible_no_itn` | --no-itn only supported by qwen3-asr-flash |
| `incompatible_vocabulary_id` | --vocabulary-id not supported by this model |
| `incompatible_disfluency_removal` | --disfluency-removal not supported by this model |
| `incompatible_diarize` | --diarize not supported by this model |
| `incompatible_itn` | --itn only supported by qwen3-asr-flash-filetrans |
| `incompatible_words` | --words only supported by qwen3-asr-flash-filetrans |
| `speakers_requires_diarize` | --speakers requires --diarize |
| `output_write_error` | Cannot write to output file |

### API Errors

| Code | Description |
|------|-------------|
| `invalid_api_key` | API key is invalid or region mismatch |
| `invalid_request` | Invalid request parameters |
| `task_not_found` | Task ID not found or expired (24h) |
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
| `qwen3-asr-flash` (no realtime/filetrans) | HTTP sync | `POST {base_url_compatible}/chat/completions` |
| `paraformer-realtime-*`, `fun-asr-realtime*` | WebSocket (run-task) | `wss://dashscope.aliyuncs.com/api-ws/v1/inference/` |
| `qwen3-asr-flash-realtime*` | WebSocket (session.update) | `wss://dashscope.aliyuncs.com/api-ws/v1/realtime?model={model}` |
| `paraformer-*`, `fun-asr*`, `qwen3-asr-flash-filetrans*` (non-realtime) | HTTP async | `POST {base_url}/services/audio/asr/transcription` |

### HTTP Sync Flow (qwen3-asr-flash)

1. Read audio file → base64 encode
2. POST to `/compatible-mode/v1/chat/completions` with base64 audio in messages
3. Parse response → extract text, language, emotion
4. Output JSON result

### WebSocket Flow — run-task (paraformer/fun-asr realtime)

1. Connect to `wss://...api-ws/v1/inference/` with auth header
2. Send `run-task` JSON (model, format, sample_rate, parameters)
3. Wait for `task-started` event
4. Stream audio file as binary frames (~100ms chunks)
5. Receive `result-generated` events (text, timestamps, words)
6. Send `finish-task` JSON
7. Wait for `task-finished` event
8. Aggregate sentences → output JSON result

### WebSocket Flow — session.update (qwen3-asr-flash-realtime)

1. Connect to `wss://...api-ws/v1/realtime?model={model}` with auth header + `OpenAI-Beta: realtime=v1`
2. Receive `session.created`
3. Send `session.update` (input_audio_format, sample_rate, language)
4. Receive `session.updated`
5. Send `input_audio_buffer.append` events (base64-encoded audio chunks)
6. Receive transcription result events
7. Send `session.finish`
8. Wait for `session.finished`
9. Output JSON result

### Async REST Flow (create/status)

1. POST to `/services/audio/asr/transcription` with `X-DashScope-Async: enable`
2. Receive `task_id` + `PENDING` status
3. Poll `GET /tasks/{task_id}` until `SUCCEEDED` or `FAILED`
4. Download transcription JSON from `transcription_url`
5. Parse and output result
