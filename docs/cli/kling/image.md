# Kling Image Commands

Commands for Kling image generation tasks.

---

## `kling image create`

Create an image generation task.

### Usage

```bash
# Text-only
kling image create "A neon-lit street at night"

# With reference image (image-to-image)
kling image create "Make it cinematic" --image ./input.png

# Use image reference type (kling-v1-5 only)
kling image create "Portrait, high detail" --image ./face.jpg --image-reference subject --human-fidelity 0.6
```

### Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `--image` | | string | | No | Reference image (local file or URL) |
| `--image-reference` | | string | | No | `subject` or `face` (kling-v1-5 only) |
| `--image-fidelity` | | float | 0.5 | No | Image reference strength (0-1) |
| `--human-fidelity` | | float | 0.45 | No | Face reference strength (0-1, requires `subject`) |
| `--negative` | | string | | No | Negative prompt (not supported with `--image`) |
| `--model` | `-m` | string | `kling-v1` | No | Model name |
| `--resolution` | | string | `1k` | No | Resolution: `1k`, `2k` |
| `--count` | `-n` | int | 1 | No | Number of images (1-9) |
| `--ratio` | `-r` | string | `16:9` | No | Aspect ratio |
| `--watermark` | | bool | false | No | Include watermark |
| `--prompt-file` | `-f` | string | | No | Read prompt from file |
| `--callback-url` | | string | | No | Callback URL |
| `--external-task-id` | | string | | No | External task ID |

### Output

```json
{
  "success": true,
  "task_id": "xxx",
  "status": "submitted"
}
```

---

## `kling image status`

Get image generation status.

### Usage

```bash
kling image status <task_id>
kling image status <task_id> -v
```

---

## `kling image download`

Download a completed image.

### Usage

```bash
kling image download <task_id> -o output.png
kling image download <task_id> -o output.png --index 1
kling image download <task_id> -o output.png --watermark
```

---

## `kling image list`

List image generation tasks.

### Usage

```bash
kling image list
kling image list --limit 50 --page 2
```
