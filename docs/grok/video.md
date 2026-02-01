# grok video

Video generation and editing using xAI Grok API.

Video operations are asynchronous. Use `create` or `edit` to start a job, then `status` to check progress, and `download` to retrieve the result.

## Commands

### grok video create

Create a video generation job.

```bash
# Text-to-video
rawgenai grok video create "A dancing cat"

# Image-to-video
rawgenai grok video create "Animate this image" -i photo.png

# With options
rawgenai grok video create "A sunset timelapse" -d 10 -a 9:16 -r 480p

# From file
rawgenai grok video create --prompt-file prompt.txt
```

#### Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--prompt-file` | | string | | Read prompt from file |
| `--image` | `-i` | string | | Input image (image-to-video) |
| `--duration` | `-d` | int | 5 | Duration in seconds (1-15) |
| `--aspect` | `-a` | string | "16:9" | Aspect ratio: 16:9, 9:16 |
| `--resolution` | `-r` | string | "720p" | Resolution: 720p, 480p |

#### Output

```json
{
  "success": true,
  "request_id": "req_abc123xyz",
  "status": "pending"
}
```

---

### grok video edit

Edit an existing video.

```bash
rawgenai grok video edit "Make it faster" -v https://example.com/video.mp4
```

#### Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--video` | `-v` | string | (required) | Input video URL |
| `--prompt-file` | | string | | Read prompt from file |

#### Output

```json
{
  "success": true,
  "request_id": "req_xyz789abc",
  "status": "pending"
}
```

---

### grok video status

Check the status of a video generation job.

```bash
rawgenai grok video status req_abc123xyz
```

#### Output

```json
{
  "success": true,
  "request_id": "req_abc123xyz",
  "status": "completed",
  "progress": 1.0
}
```

Possible status values:
- `pending` - Job queued
- `running` - Generation in progress
- `completed` / `succeeded` - Ready to download
- `failed` - Generation failed

---

### grok video download

Download a completed video.

```bash
rawgenai grok video download req_abc123xyz -o video.mp4
```

#### Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--output` | `-o` | string | (required) | Output file path (.mp4) |

#### Output

```json
{
  "success": true,
  "request_id": "req_abc123xyz",
  "file": "/absolute/path/to/video.mp4"
}
```

---

## Error Codes

### CLI Errors

| Code | Description |
|------|-------------|
| `missing_prompt` | No prompt provided |
| `missing_output` | No output file specified |
| `missing_video` | No video URL specified (edit) |
| `missing_request_id` | No request ID provided |
| `missing_api_key` | XAI_API_KEY not set |
| `invalid_duration` | Duration not in 1-15 range |
| `invalid_aspect` | Invalid aspect ratio |
| `invalid_resolution` | Invalid resolution |
| `invalid_format` | Output format not .mp4 |
| `image_not_found` | Input image not found |
| `invalid_image_format` | Unsupported image format |
| `video_not_ready` | Video not ready for download |
| `video_failed` | Video generation failed |

### API Errors

| Code | Description |
|------|-------------|
| `invalid_request` | Bad request (400) |
| `invalid_api_key` | Invalid or revoked API key (401) |
| `permission_denied` | Insufficient permissions (403) |
| `not_found` | Request ID not found (404) |
| `rate_limit` | Too many requests (429) |
| `quota_exceeded` | API quota exhausted |
| `server_error` | xAI server error (500) |
| `server_overloaded` | Server overloaded (503) |
| `timeout` | Request timed out |
| `connection_error` | Cannot connect to API |
| `download_error` | Failed to download video |

## Environment Variables

| Variable | Description |
|----------|-------------|
| `XAI_API_KEY` | xAI API key (required) |

## Workflow Example

```bash
# 1. Create a video
rawgenai grok video create "A flying eagle over mountains" -d 10
# Output: {"success":true,"request_id":"req_abc123","status":"pending"}

# 2. Check status (poll until completed)
rawgenai grok video status req_abc123
# Output: {"success":true,"request_id":"req_abc123","status":"running","progress":0.5}

# 3. Download when completed
rawgenai grok video download req_abc123 -o eagle.mp4
# Output: {"success":true,"request_id":"req_abc123","file":"/path/to/eagle.mp4"}
```
