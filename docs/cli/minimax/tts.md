# MiniMax TTS

MiniMax 文本转语音（同步、异步长文本、WebSocket 流式）。

## Usage

```bash
rawgenai minimax tts [text] [flags]
rawgenai minimax tts --stream [text] [flags]
rawgenai minimax tts create [text] [flags]
rawgenai minimax tts status <task_id>
rawgenai minimax tts download <file_id> -o out.mp3
```

## Flags（同步/流式）

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--output` | `-o` | string | | 输出文件路径（不带 `--speak` 时必填） |
| `--prompt-file` | | string | | 从文件读取文本 |
| `--model` | `-m` | string | `speech-2.8-hd` | 模型名称 |
| `--voice` | | string | `English_Graceful_Lady` | Voice ID |
| `--speed` | | float | `1` | 语速 `0.5-2.0` |
| `--vol` | | float | `1` | 音量 `(0,10]` |
| `--pitch` | | int | `0` | 音高 `-12~12` |
| `--format` | | string | `mp3` | 音频格式：`mp3`/`pcm`/`flac`/`wav` |
| `--sample-rate` | | int | `0` | 采样率（可选） |
| `--bitrate` | | int | `0` | 比特率（mp3 可用） |
| `--channel` | | int | `0` | 声道数 `1/2` |
| `--stream` | | bool | `false` | 使用 WebSocket 流式 |
| `--speak` | | bool | `false` | 生成后播放（仅支持 `mp3`） |

## Flags（异步 create）

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--prompt-file` | | string | | 从文件读取文本 |
| `--model` | `-m` | string | `speech-2.8-hd` | 模型名称 |
| `--voice` | | string | `English_Graceful_Lady` | Voice ID |
| `--speed` | | float | `1` | 语速 `0.5-2.0` |
| `--vol` | | float | `1` | 音量 `(0,10]` |
| `--pitch` | | int | `0` | 音高 `-12~12` |
| `--format` | | string | `mp3` | 音频格式：`mp3`/`pcm`/`flac`/`wav` |
| `--sample-rate` | | int | `0` | 采样率（可选） |
| `--bitrate` | | int | `0` | 比特率（mp3 可用） |
| `--channel` | | int | `0` | 声道数 `1/2` |
| `--file-id` | | int64 | `0` | 长文本文件 ID（与文本互斥） |

## Environment Variables

- `MINIMAX_API_KEY` - MiniMax API Key（必填）

## Examples

同步：
```bash
rawgenai minimax tts "你好世界" -o out.mp3
```

同步 + 播放：
```bash
rawgenai minimax tts "Hello" --speak
```

WebSocket 流式：
```bash
rawgenai minimax tts "实时播报一下" --stream -o out.mp3
```

异步长文本：
```bash
rawgenai minimax tts create "很长的文本..." --model speech-2.8-hd
rawgenai minimax tts status <task_id>
rawgenai minimax tts download <file_id> -o out.mp3
```

## Output Format

```json
{
  "success": true,
  "file": "/path/to/output.mp3",
  "model": "speech-2.8-hd",
  "voice": "English_Graceful_Lady"
}
```

## Error Codes

| Code | Description |
|------|-------------|
| `missing_text` | 未提供文本 |
| `missing_output` | 未提供输出文件（且未使用 `--speak`） |
| `invalid_format` | 音频格式不支持 |
| `invalid_speed` | 语速超出范围 |
| `invalid_volume` | 音量超出范围 |
| `invalid_pitch` | 音高超出范围 |
| `invalid_sample_rate` | 采样率不合法 |
| `invalid_bitrate` | 比特率不合法 |
| `invalid_channel` | 声道数不合法 |
| `missing_api_key` | 未设置 `MINIMAX_API_KEY` |
| `api_error` | API 返回错误 |
| `decode_error` | 音频解码失败 |
| `output_write_error` | 输出写入失败 |
| `playback_error` | 播放失败 |
