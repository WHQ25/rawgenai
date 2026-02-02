# MiniMax Image

MiniMax 生图（文生图 / 图生图）。

## Usage

```bash
rawgenai minimax image create [prompt] [flags]
```

## Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--output` | `-o` | string | | 输出文件路径（base64 必填，url 可选） |
| `--prompt-file` | | string | | 从文件读取提示词 |
| `--image` | `-i` | string[] | | 参考图（可重复，用于图生图） |
| `--model` | `-m` | string | `image-01` | 模型：`image-01`、`image-01-live`（需 `--image`） |
| `--aspect` | | string | `1:1` | 画面比例 |
| `--width` | | int | `0` | 宽度（512-2048，8 的倍数） |
| `--height` | | int | `0` | 高度（512-2048，8 的倍数） |
| `--n` | `-n` | int | `1` | 生成张数（1-9） |
| `--response-format` | | string | `base64` | `base64` 或 `url` |
| `--prompt-optimizer` | | bool | `false` | 启用提示词优化 |

## Environment Variables

- `MINIMAX_API_KEY` - MiniMax API Key（必填）

## Examples

文生图：
```bash
rawgenai minimax image create "a cat in a park" -o out.png
```

图生图（主体参考）：
```bash
rawgenai minimax image create "cinematic portrait" \
  --image ./face.png \
  -o out.png
```

使用 live 模型（需要参考图）：
```bash
rawgenai minimax image create "anime style" \
  --model image-01-live \
  --image ./face.png \
  -o out.png
```

多张输出：
```bash
rawgenai minimax image create "a cat" -n 3 -o out.png
```

只获取 URL（不下载）：
```bash
rawgenai minimax image create "a cat" --response-format url
```

## Output Format

下载到本地：
```json
{
  "success": true,
  "file": "/path/to/out.png",
  "model": "image-01",
  "count": 1
}
```

只返回 URL（`--response-format url` 且无 `-o`）：
```json
{
  "success": true,
  "url": "https://...",
  "model": "image-01",
  "count": 1
}
```

多张图片：
```json
{
  "success": true,
  "files": ["/path/to/out_1.png", "/path/to/out_2.png"],
  "model": "image-01",
  "count": 2
}
```

## Error Codes

| Code | Description |
|------|-------------|
| `missing_prompt` | 未提供提示词 |
| `missing_output` | 未提供输出文件（base64 模式） |
| `invalid_model` | 模型不合法 |
| `invalid_aspect` | 画面比例不合法 |
| `invalid_count` | 张数不合法 |
| `invalid_response_format` | 返回格式不合法 |
| `missing_image` | `image-01-live` 未提供参考图 |
| `unsupported_format` | 输出格式不支持 |
| `missing_api_key` | 未设置 `MINIMAX_API_KEY` |
| `api_error` | API 返回错误 |
| `output_write_error` | 输出写入失败 |
