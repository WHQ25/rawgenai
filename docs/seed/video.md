# rawgenai seed video

Generate videos using ByteDance Seedance 1.5 Pro model.

## Commands

| Command | Description |
|---------|-------------|
| `create` | Create a video generation task |
| `status` | Get video generation status |
| `download` | Download completed video |
| `list` | List video generation tasks |
| `delete` | Delete/cancel a task |

## seed video create

Create a video generation task.

### Usage

```bash
rawgenai seed video create <prompt> [flags]
rawgenai seed video create --prompt-file <file> [flags]
cat prompt.txt | rawgenai seed video create [flags]
```

### Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `--prompt-file` | - | string | - | No | Input prompt file |
| `--first-frame` | - | string | - | No | First frame image (JPEG/PNG/WebP) |
| `--last-frame` | - | string | - | No | Last frame image (requires --first-frame) |
| `--ratio` | `-r` | string | `16:9` | No | Aspect ratio: 16:9, 9:16, 4:3, 3:4, 1:1, 21:9 |
| `--resolution` | - | string | `1080p` | No | Resolution: 480p, 720p, 1080p |
| `--duration` | `-d` | int | `5` | No | Duration in seconds (4-12) |
| `--audio` | - | bool | `false` | No | Generate video with audio |
| `--seed` | - | int | - | No | Random seed for reproducibility |
| `--watermark` | - | bool | `false` | No | Add watermark to output |
| `--return-last-frame` | - | bool | `false` | No | Return last frame URL (for chaining) |

### Examples

```bash
# Text to video
rawgenai seed video create "一只猫在钢琴上弹奏"

# With first frame (image to video)
rawgenai seed video create "女孩转头微笑" --first-frame photo.jpg

# With first and last frame (interpolation)
rawgenai seed video create "镜头从A移动到B" --first-frame start.jpg --last-frame end.jpg

# With audio
rawgenai seed video create "海浪拍打沙滩的声音" --audio

# Custom settings
rawgenai seed video create "竖屏短视频" --ratio 9:16 --resolution 720p --duration 8
```

### Output

```json
{
  "success": true,
  "task_id": "cgt-2025xxxx-xxxx",
  "status": "queued"
}
```

## seed video status

Get video generation status.

### Usage

```bash
rawgenai seed video status <task_id>
```

### Output

**Queued/Running:**
```json
{
  "success": true,
  "task_id": "cgt-2025xxxx-xxxx",
  "status": "running"
}
```

**Succeeded:**
```json
{
  "success": true,
  "task_id": "cgt-2025xxxx-xxxx",
  "status": "succeeded",
  "video_url": "https://...",
  "last_frame_url": "https://...",
  "resolution": "1080p",
  "ratio": "16:9",
  "duration": 5,
  "seed": 58944
}
```

**Failed:**
```json
{
  "success": false,
  "error": {
    "code": "content_policy",
    "message": "Content violates safety policy"
  }
}
```

### Status Values

| Status | Description |
|--------|-------------|
| `queued` | Task is waiting in queue |
| `running` | Task is being processed |
| `succeeded` | Task completed successfully |
| `failed` | Task failed |

## seed video download

Download completed video.

### Usage

```bash
rawgenai seed video download <task_id> -o <output.mp4>
```

### Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `--output` | `-o` | string | - | Yes | Output file path (.mp4) |
| `--last-frame` | - | string | - | No | Also save last frame to this path |

### Examples

```bash
# Download video
rawgenai seed video download cgt-2025xxxx -o video.mp4

# Download video and last frame (for chaining)
rawgenai seed video download cgt-2025xxxx -o video.mp4 --last-frame last.jpg
```

### Output

```json
{
  "success": true,
  "task_id": "cgt-2025xxxx-xxxx",
  "file": "/path/to/video.mp4",
  "last_frame_file": "/path/to/last.jpg"
}
```

## seed video list

List video generation tasks.

### Usage

```bash
rawgenai seed video list [flags]
```

### Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `--limit` | `-l` | int | `20` | No | Maximum number of tasks to return (1-100) |
| `--status` | `-s` | string | - | No | Filter by status: queued, running, succeeded, failed |

### Examples

```bash
# List recent tasks
rawgenai seed video list

# List more tasks
rawgenai seed video list --limit 50

# List only running tasks
rawgenai seed video list --status running

# List completed tasks
rawgenai seed video list --status succeeded
```

### Output

```json
{
  "success": true,
  "tasks": [
    {
      "task_id": "cgt-2025xxxx-xxxx",
      "status": "succeeded",
      "created_at": 1765510475
    },
    {
      "task_id": "cgt-2025xxxx-yyyy",
      "status": "running",
      "created_at": 1765510400
    }
  ],
  "count": 2
}
```

## seed video delete

Delete or cancel a video generation task.

**Note:** Only `queued` tasks can be cancelled. Tasks in `running` status cannot be cancelled or deleted until they complete.

### Usage

```bash
rawgenai seed video delete <task_id>
```

### Allowed Operations

| Status | Action | Allowed |
|--------|--------|---------|
| `queued` | Cancel | Yes |
| `running` | Cancel | No |
| `succeeded` | Delete | Yes |
| `failed` | Delete | Yes |

### Examples

```bash
# Cancel a queued task
rawgenai seed video delete cgt-2025xxxx-xxxx

# Delete a completed task
rawgenai seed video delete cgt-2025xxxx-yyyy
```

### Output

```json
{
  "success": true,
  "task_id": "cgt-2025xxxx-xxxx",
  "deleted": true
}
```

## Environment Variables

- `ARK_API_KEY` - ByteDance Ark API key (required)

## Model

| Model | Model ID | Features |
|-------|----------|----------|
| Seedance 1.5 Pro | `doubao-seedance-1-5-pro-251215` | Best quality, audio support, first/last frame |

## Constraints

### Duration
- Range: 4-12 seconds

### Resolution & Ratio

| Resolution | Supported Ratios |
|------------|------------------|
| 480p | 16:9, 9:16, 4:3, 3:4, 1:1, 21:9 |
| 720p | 16:9, 9:16, 4:3, 3:4, 1:1, 21:9 |
| 1080p | 16:9, 9:16, 4:3, 3:4, 1:1, 21:9 |

### First/Last Frame
- Last frame requires first frame
- Images are auto-cropped to match ratio
- Upload high-quality images for best results

### Task Expiration
- Task data expires after 24 hours
- Download videos promptly

## Errors

### CLI Errors

| Code | Description |
|------|-------------|
| `missing_api_key` | ARK_API_KEY not set |
| `missing_prompt` | No prompt provided |
| `missing_task_id` | Task ID not provided |
| `missing_output` | Output file required |
| `file_not_found` | Input file not found |
| `first_frame_not_found` | First frame image not found |
| `last_frame_not_found` | Last frame image not found |
| `last_frame_requires_first` | --last-frame requires --first-frame |
| `invalid_format` | Invalid output format |
| `invalid_ratio` | Invalid aspect ratio |
| `invalid_resolution` | Invalid resolution |
| `invalid_duration` | Duration out of range (4-12) |
| `invalid_limit` | Limit out of range (1-100) |
| `invalid_status` | Invalid status filter |
| `image_read_error` | Cannot read image file |
| `output_write_error` | Cannot write output file |

### API Errors

| Code | Description |
|------|-------------|
| `invalid_request` | Invalid request parameters |
| `content_policy` | Content violates safety policy |
| `invalid_api_key` | API key is invalid |
| `permission_denied` | API key lacks permissions |
| `task_not_found` | Task ID not found |
| `rate_limit` | Too many requests |
| `quota_exceeded` | Quota exhausted |
| `server_error` | Server error |

### Download Errors

| Code | Description |
|------|-------------|
| `video_not_ready` | Video generation not completed |
| `video_failed` | Video generation failed |
| `no_video` | No video URL in response |
| `download_error` | Cannot download video |
| `connection_error` | Network connection failed |
| `timeout` | Request timed out |
