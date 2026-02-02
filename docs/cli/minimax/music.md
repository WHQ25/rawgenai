# MiniMax Music

MiniMax 音乐生成（歌词生成音乐）。

## Usage

```bash
rawgenai minimax music create [lyrics] [flags]
```

## Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--output` | `-o` | string | | 输出文件路径 |
| `--lyrics-file` | | string | | 从文件读取歌词 |
| `--prompt` | `-p` | string | | 音乐风格描述 |
| `--play` | | bool | `false` | 生成后播放 |
| `--stream` | | bool | `false` | 流式输出到 stdout |
| `--format` | `-f` | string | `mp3` | 音频格式：`mp3` / `wav` / `pcm` |
| `--sample-rate` | | int | `44100` | 采样率：16000/24000/32000/44100 |
| `--bitrate` | | int | `256000` | 比特率：32000/64000/128000/256000 |

## 歌词结构标签

支持以下结构标签来增强编曲：

`[Intro]`, `[Verse]`, `[Pre Chorus]`, `[Chorus]`, `[Interlude]`, `[Bridge]`, `[Outro]`, `[Post Chorus]`, `[Transition]`, `[Break]`, `[Hook]`, `[Build Up]`, `[Inst]`, `[Solo]`

## Environment Variables

- `MINIMAX_API_KEY` - MiniMax API Key（必填）

## Examples

基本生成：
```bash
rawgenai minimax music create "[verse]\n月光洒落窗前\n思念悄悄蔓延" \
  -p "Pop, melancholic, piano" \
  -o song.mp3
```

从文件读取歌词：
```bash
rawgenai minimax music create --lyrics-file lyrics.txt -p "Rock, energetic" -o rock.mp3
```

直接播放（不保存）：
```bash
rawgenai minimax music create "[chorus]\nLet it go" -p "Pop, uplifting" --play
```

保存并播放：
```bash
rawgenai minimax music create "[verse]\nHello world" -o hello.mp3 --play
```

流式输出：
```bash
rawgenai minimax music create "[verse]\nStreaming" --stream > output.mp3
```

## Output Format

```json
{
  "success": true,
  "file": "/path/to/song.mp3",
  "duration_ms": 180000,
  "size_bytes": 2880000
}
```

## Error Codes

| Code | Description |
|------|-------------|
| `missing_lyrics` | 未提供歌词 |
| `lyrics_too_long` | 歌词超过 3500 字符 |
| `missing_output` | 未提供输出文件（非 --play/--stream 模式） |
| `invalid_format` | 音频格式不合法 |
| `invalid_sample_rate` | 采样率不合法 |
| `invalid_bitrate` | 比特率不合法 |
| `format_mismatch` | 输出文件扩展名与格式不匹配 |
| `missing_api_key` | 未设置 `MINIMAX_API_KEY` |
| `api_error` | API 返回错误 |
| `playback_error` | 播放失败 |
