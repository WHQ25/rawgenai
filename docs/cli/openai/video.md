# rawgenai openai video

Generate video using OpenAI Sora models.

## Commands

| Command | Description |
|---------|-------------|
| `create` | Create a video generation job |
| `status` | Get video generation status |
| `download` | Download video content (video/thumbnail/spritesheet) |
| `list` | List all video generation jobs |
| `delete` | Delete a video |
| `remix` | Create a remix from an existing video |

---

## video create

Create a video generation job. Returns immediately with a video ID for async workflow.

### Usage

```bash
rawgenai openai video create <prompt> [flags]
rawgenai openai video create --file <prompt.txt> [flags]
cat prompt.txt | rawgenai openai video create [flags]
```

### Examples

```bash
# Basic generation
rawgenai openai video create "A cat playing piano on stage"

# High quality, longer duration
rawgenai openai video create "A sunset over the ocean with waves" --model sora-2-pro --duration 8

# Portrait video (for mobile)
rawgenai openai video create "A person walking in the rain" --size 720x1280

# Wide cinematic shot
rawgenai openai video create "Aerial view of a mountain range" --size 1792x1024

# With first frame image (image-to-video)
rawgenai openai video create "She turns around and smiles" --image first_frame.jpg

# From file
rawgenai openai video create --file prompt.txt

# From stdin
echo "A dog running in a park" | rawgenai openai video create
```

### Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `--file` | - | string | - | No | Input prompt file |
| `--image` | `-i` | string | - | No | First frame image (JPEG/PNG/WebP) |
| `--model` | `-m` | string | `sora-2` | No | Model name |
| `--size` | `-s` | string | `1280x720` | No | Video resolution |
| `--duration` | `-d` | int | `4` | No | Video duration in seconds (4, 8, 12) |

### Output

```json
{
  "success": true,
  "video_id": "video_abc123",
  "status": "queued",
  "model": "sora-2",
  "size": "1280x720",
  "duration": 4,
  "created_at": 1706745600
}
```

---

## video status

Query the status of a video generation job.

### Usage

```bash
rawgenai openai video status <video_id>
```

### Examples

```bash
# Check status
rawgenai openai video status video_abc123
```

### Output

```json
{
  "success": true,
  "video_id": "video_abc123",
  "status": "completed",
  "created_at": 1706745600
}
```

#### Status Values

| Status | Description |
|--------|-------------|
| `queued` | Job is waiting to be processed |
| `in_progress` | Video is being generated |
| `completed` | Video is ready for download |
| `failed` | Generation failed (see error_message) |

#### Failed Status

```json
{
  "success": true,
  "video_id": "video_abc123",
  "status": "failed",
  "error_message": "Content policy violation",
  "created_at": 1706745600
}
```

---

## video download

Download video content from a completed video job.

### Usage

```bash
rawgenai openai video download <video_id> -o <output_file> [flags]
```

### Examples

```bash
# Download video
rawgenai openai video download video_abc123 -o my_video.mp4

# Download thumbnail
rawgenai openai video download video_abc123 -o thumbnail.jpg --variant thumbnail

# Download spritesheet
rawgenai openai video download video_abc123 -o sprite.jpg --variant spritesheet
```

### Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `--output` | `-o` | string | - | Yes | Output file path |
| `--variant` | - | string | `video` | No | Content type: video, thumbnail, spritesheet |

### Variant File Extensions

| Variant | Extension |
|---------|-----------|
| `video` | `.mp4` |
| `thumbnail` | `.jpg` |
| `spritesheet` | `.jpg` |

### Output

```json
{
  "success": true,
  "video_id": "video_abc123",
  "variant": "video",
  "file": "/path/to/my_video.mp4"
}
```

---

## video list

List all video generation jobs.

### Usage

```bash
rawgenai openai video list [flags]
```

### Examples

```bash
# List recent videos
rawgenai openai video list

# List more videos
rawgenai openai video list --limit 50

# List oldest first
rawgenai openai video list --order asc
```

### Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `--limit` | `-l` | int | `20` | No | Maximum number of results |
| `--order` | - | string | `desc` | No | Sort order: asc, desc |

### Output

```json
{
  "success": true,
  "videos": [
    {
      "video_id": "video_abc123",
      "status": "completed",
      "created_at": 1706745600
    },
    {
      "video_id": "video_def456",
      "status": "in_progress",
      "created_at": 1706745500
    }
  ],
  "count": 2
}
```

---

## video delete

Delete a video and its associated assets.

### Usage

```bash
rawgenai openai video delete <video_id>
```

### Examples

```bash
# Delete a video
rawgenai openai video delete video_abc123
```

### Output

```json
{
  "success": true,
  "video_id": "video_abc123",
  "deleted": true
}
```

---

## video remix

Create a new video based on an existing video with a new prompt.

### Usage

```bash
rawgenai openai video remix <video_id> <prompt> [flags]
rawgenai openai video remix <video_id> --file <prompt.txt>
cat prompt.txt | rawgenai openai video remix <video_id>
```

### Examples

```bash
# Remix with new prompt
rawgenai openai video remix video_abc123 "Make it night time with stars"

# Remix from file
rawgenai openai video remix video_abc123 --file new_prompt.txt
```

### Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `--file` | - | string | - | No | Input prompt file |

### Output

```json
{
  "success": true,
  "video_id": "video_new789",
  "status": "queued",
  "remixed_from_id": "video_abc123",
  "created_at": 1706745700
}
```

---

## Workflow Example

Video generation is asynchronous. Use the following workflow:

```bash
# 1. Create generation job (returns immediately)
rawgenai openai video create "A beautiful sunset"
# Output: {"success":true,"video_id":"video_abc123","status":"queued",...}

# 2. Check status periodically
rawgenai openai video status video_abc123
# Output: {"success":true,"video_id":"video_abc123","status":"in_progress",...}

# 3. When completed, download the video
rawgenai openai video download video_abc123 -o sunset.mp4
# Output: {"success":true,"video_id":"video_abc123","variant":"video","file":"/path/to/sunset.mp4"}

# Optional: Download thumbnail
rawgenai openai video download video_abc123 -o sunset_thumb.jpg --variant thumbnail

# Optional: List all your videos
rawgenai openai video list

# Optional: Create a remix
rawgenai openai video remix video_abc123 "Same scene but at night"

# Optional: Delete when done
rawgenai openai video delete video_abc123
```

---

## Models

| Model | Description |
|-------|-------------|
| `sora-2` | Fast, flexible video generation (default) |
| `sora-2-pro` | Production quality, higher fidelity |

## Sizes

| Size | Aspect Ratio | Description |
|------|--------------|-------------|
| `1280x720` | 16:9 | Landscape HD (default) |
| `720x1280` | 9:16 | Portrait (mobile) |
| `1792x1024` | ~16:9 | Wide cinematic |
| `1024x1792` | ~9:16 | Tall portrait |

## Duration

| Seconds | Description |
|---------|-------------|
| `4` | Short clip (default) |
| `8` | Medium length |
| `12` | Longer clip |

## First Frame Image

Use `--image` to provide a reference image as the first frame of the video. This is useful for:
- Preserving brand assets or characters
- Maintaining specific environments
- Image-to-video generation

**Requirements:**
- Image resolution must match `--size` parameter
- Supported formats: JPEG, PNG, WebP
- Images with human faces are rejected

---

## Errors

```json
{
  "success": false,
  "error": {
    "code": "video_not_ready",
    "message": "video is not ready for download, current status: in_progress"
  }
}
```

### CLI Errors (before API call)

| Code | Description |
|------|-------------|
| `missing_api_key` | OPENAI_API_KEY not set |
| `missing_prompt` | No prompt provided |
| `missing_output` | --output flag not provided |
| `missing_video_id` | video_id argument not provided |
| `file_not_found` | Input file does not exist |
| `image_not_found` | Image file does not exist |
| `invalid_image_format` | Image format not supported |
| `invalid_format` | Output file extension is incorrect |
| `invalid_size` | Size value not allowed |
| `invalid_duration` | Duration value not allowed (4, 8, 12) |
| `invalid_variant` | Variant must be: video, thumbnail, spritesheet |
| `invalid_limit` | Limit must be between 1 and 100 |
| `invalid_order` | Order must be: asc, desc |
| `output_write_error` | Cannot write to output file |

### OpenAI API Errors

| HTTP | Code | Description |
|------|------|-------------|
| 400 | `invalid_model` | Model does not exist |
| 400 | `content_policy` | Content violates usage policies |
| 400 | `invalid_request` | Other invalid request parameters |
| 401 | `invalid_api_key` | API key is invalid or revoked |
| 403 | `region_not_supported` | Region/country not supported |
| 404 | `video_not_found` | Video ID does not exist |
| 429 | `rate_limit` | Too many requests |
| 429 | `quota_exceeded` | Quota/credits exhausted |
| 500 | `server_error` | OpenAI server error |
| 503 | `server_overloaded` | OpenAI server overloaded |

### Download Errors

| Code | Description |
|------|-------------|
| `video_not_ready` | Video is not completed yet |

### Network Errors

| Code | Description |
|------|-------------|
| `connection_error` | Cannot connect to OpenAI API |
| `timeout` | Request timed out |

---

## Prompt Tips

For best results, describe:
- **Shot type**: wide, close-up, tracking, aerial
- **Subject**: what/who is in the video
- **Action**: what is happening
- **Setting**: where it takes place
- **Lighting**: natural, dramatic, soft, etc.

Example: "Wide tracking shot of a teal coupe driving through a desert highway, heat ripples visible, hard sun overhead."

## Content Restrictions

- Content must be suitable for audiences under 18
- Copyrighted characters and music will be rejected
- Real people and public figures cannot be generated
- Input images with human faces are currently rejected
