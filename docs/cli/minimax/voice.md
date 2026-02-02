# MiniMax Voice

MiniMax 声音管理（列表、上传、克隆、声音设计、删除）。

## Usage

```bash
rawgenai minimax voice list [flags]
rawgenai minimax voice upload --file ./sample.mp3
rawgenai minimax voice clone --file-id 123 --voice-id my_voice
rawgenai minimax voice design --prompt "..." --preview-text "..." -o trial.mp3
rawgenai minimax voice delete <voice_id> --type voice_cloning
```

## 子命令与 Flags

### list
| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--type` | `-t` | string | `all` | 取值：`all`/`system`/`voice_cloning`/`voice_generation` |

### upload
| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--file` | `-f` | string | | 音频文件路径（必填） |
| `--purpose` | | string | `voice_clone` | 目的（当前仅支持 `voice_clone`） |

### clone
| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--file-id` | | int64 | `0` | 上传音频文件 ID（必填） |
| `--voice-id` | | string | | 新的 voice_id（必填） |
| `--prompt-audio-id` | | int64 | `0` | prompt 音频文件 ID（可选） |
| `--prompt-text` | | string | | prompt 音频对应文本（与上面的 ID 成对出现） |
| `--preview-text` | | string | | 生成 demo 音频的文本 |
| `--model` | `-m` | string | `speech-2.8-hd` | demo 生成模型 |
| `--language-boost` | | string | | 语言增强（可选） |
| `--noise-reduction` | | bool | `false` | 降噪 |
| `--volume-normalization` | | bool | `false` | 音量归一化 |
| `--continuous-sound` | | bool | `false` | 平滑衔接 |

### design
| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--prompt` | | string | | 声音描述（必填） |
| `--prompt-file` | | string | | 从文件读取描述 |
| `--preview-text` | | string | | 试听文本（必填） |
| `--preview-file` | | string | | 从文件读取试听文本 |
| `--voice-id` | | string | | 复用已有 voice_id（可选） |
| `--output` | `-o` | string | | 试听音频输出路径（必填，除非 `--speak`） |
| `--speak` | | bool | `false` | 生成后播放（mp3） |

### delete
| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--type` | `-t` | string | `voice_cloning` | 取值：`voice_cloning`/`voice_generation` |

## Environment Variables

- `MINIMAX_API_KEY` - MiniMax API Key（必填）

## Examples

列出所有声音：
```bash
rawgenai minimax voice list
```

上传克隆音频：
```bash
rawgenai minimax voice upload --file ./sample.mp3
```

克隆声音并生成 demo：
```bash
rawgenai minimax voice clone --file-id 123 --voice-id my_voice \
  --preview-text "Hello world" --model speech-2.8-hd
```

声音设计并试听：
```bash
rawgenai minimax voice design \
  --prompt "warm, clear, professional narrator" \
  --preview-text "Welcome to our podcast." \
  -o trial.mp3
```

删除声音：
```bash
rawgenai minimax voice delete my_voice --type voice_cloning
```

## Error Codes

| Code | Description |
|------|-------------|
| `missing_api_key` | 未设置 `MINIMAX_API_KEY` |
| `invalid_type` | 类型不合法 |
| `missing_file` | 未提供音频文件 |
| `file_not_found` | 文件不存在 |
| `missing_file_id` | 缺少 file-id |
| `missing_voice_id` | 缺少 voice-id |
| `invalid_prompt` | prompt 音频/文本不匹配 |
| `missing_prompt` | 缺少设计描述 |
| `missing_preview_text` | 缺少试听文本 |
| `missing_output` | 未提供输出文件（且未 `--speak`） |
| `api_error` | API 返回错误 |
