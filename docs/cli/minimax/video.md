# MiniMax Video

MiniMax 生视频（文生视频、图生视频、首尾帧、主体参考）。

## Usage

```bash
rawgenai minimax video create [prompt] [flags]
rawgenai minimax video status <task_id>
rawgenai minimax video download <file_id> -o out.mp4
```

## Flags（create）

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--model` | `-m` | string | 自动选择 | 模型名称 |
| `--prompt-file` | | string | | 从文件读取提示词 |
| `--duration` | `-d` | int | `6` | 时长（秒） |
| `--resolution` | `-r` | string | | 分辨率 |
| `--prompt-optimizer` | | bool | `true` | 启用提示词优化 |
| `--fast-pretreatment` | | bool | `false` | 快速预处理（Hailuo 模型） |
| `--callback-url` | | string | | 回调 URL |
| `--first-frame` | | string | | 首帧图片 |
| `--last-frame` | | string | | 末帧图片（需配合 `--first-frame`） |
| `--subject` | | string | | 主体参考图 |

## 自动类型推断

根据传入参数自动判断生成类型：

| 参数组合 | 类型 | 默认模型 |
|---------|------|---------|
| `--subject` | s2v | S2V-01（固定） |
| `--first-frame` + `--last-frame` | fl2v | MiniMax-Hailuo-02（固定） |
| `--first-frame` | i2v | MiniMax-Hailuo-2.3 |
| 无图像参数 | t2v | MiniMax-Hailuo-2.3 |

## 支持的模型

**文生视频（t2v）：**
- MiniMax-Hailuo-2.3、MiniMax-Hailuo-02、T2V-01-Director、T2V-01

**图生视频（i2v）：**
- MiniMax-Hailuo-2.3、MiniMax-Hailuo-2.3-Fast、MiniMax-Hailuo-02
- I2V-01-Director、I2V-01-live、I2V-01

**首尾帧（fl2v）：** MiniMax-Hailuo-02（固定）

**主体参考（s2v）：** S2V-01（固定）

## Environment Variables

- `MINIMAX_API_KEY` - MiniMax API Key（必填）

## Examples

文生视频：
```bash
rawgenai minimax video create "a cat running"
```

图生视频：
```bash
rawgenai minimax video create "a cat running" --first-frame ./img.png
```

首尾帧视频：
```bash
rawgenai minimax video create --first-frame ./a.png --last-frame ./b.png
```

主体参考视频：
```bash
rawgenai minimax video create "a girl smiles" --subject ./face.png
```

指定模型：
```bash
rawgenai minimax video create "a cat" -m T2V-01-Director
```

查询状态与下载：
```bash
rawgenai minimax video status <task_id>
rawgenai minimax video download <file_id> -o out.mp4
```

## Output Format

创建任务：
```json
{
  "success": true,
  "task_id": "106916112212032",
  "model": "MiniMax-Hailuo-2.3",
  "type": "t2v"
}
```

## Error Codes

| Code | Description |
|------|-------------|
| `missing_prompt` | 文生视频未提供提示词 |
| `invalid_model` | 模型不合法 |
| `invalid_resolution` | 分辨率不合法 |
| `invalid_parameter` | 参数不支持该模式（如 s2v 不支持 resolution） |
| `image_read_error` | 图片读取失败 |
| `missing_api_key` | 未设置 `MINIMAX_API_KEY` |
| `api_error` | API 返回错误 |
| `download_error` | 下载失败 |
