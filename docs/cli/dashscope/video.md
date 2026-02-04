# rawgenai dashscope video

Generate videos using Alibaba Tongyi Wanxiang (通义万象) models via DashScope API.

## Commands

| Command | Description |
|---------|-------------|
| `create` | Create a video generation task |
| `status` | Get video generation status |
| `download` | Download completed video |

> DashScope does not provide a task list API, so there is no `list` command.

## dashscope video create

Create a video generation task. Automatically selects the API based on input flags:

- Prompt only → Text-to-video (t2v)
- `--image` → Image-to-video (i2v)
- `--ref` → Reference-to-video (r2v)
- `--first-frame` → Keyframe-to-video (kf2v)

### Usage

```bash
# Text to video
rawgenai dashscope video create "一只猫在花园里玩耍"

# Image to video
rawgenai dashscope video create "让画面中的人物微笑" --image photo.jpg

# Image to video with URL
rawgenai dashscope video create "让画面动起来" --image https://example.com/photo.jpg

# Keyframe to video (first frame only)
rawgenai dashscope video create "镜头缓缓推进" --first-frame start.jpg

# Keyframe to video (first + last frame)
rawgenai dashscope video create "场景过渡" --first-frame start.jpg --last-frame end.jpg

# With audio
rawgenai dashscope video create "海浪拍打沙滩" --audio

# With custom audio file
rawgenai dashscope video create "人物说话" --image photo.jpg --audio-url https://example.com/speech.wav

# Custom settings
rawgenai dashscope video create "竖屏短视频" --ratio 9:16 --resolution 1080P --duration 10

# Multi-shot narrative (wan2.6 only)
rawgenai dashscope video create "多镜头叙事短片" --shot-type multi --duration 15

# Reference to video - single character
rawgenai dashscope video create "character1 在海边散步" --ref person.jpg

# Reference to video - multiple characters
rawgenai dashscope video create "character1 和 character2 在公园聊天" \
  --ref person1.jpg --ref person2.jpg

# Reference to video - with video reference
rawgenai dashscope video create "character1 在视频中的场景里跳舞" \
  --ref person.jpg --ref scene.mp4

# Reference to video - no audio
rawgenai dashscope video create "character1 微笑" --ref person.jpg --no-audio

# From file
rawgenai dashscope video create --prompt-file prompt.txt

# From stdin
cat prompt.txt | rawgenai dashscope video create
```

### Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `--image` | `-i` | string | - | No | Input image for i2v (local path or URL) |
| `--ref` | - | string[] | - | No | Reference file(s) for r2v (image or video, repeatable, max 5) |
| `--first-frame` | - | string | - | No | First frame image for kf2v (local path or URL) |
| `--last-frame` | - | string | - | No | Last frame image for kf2v (requires --first-frame) |
| `--prompt-file` | `-f` | string | - | No | Read prompt from file |
| `--model` | `-m` | string | auto | No | Model name (auto-selected based on input type) |
| `--resolution` | `-r` | string | `720P` | No | Resolution: 480P, 720P, 1080P |
| `--ratio` | - | string | `16:9` | No | Aspect ratio: 16:9, 9:16 (t2v/r2v only) |
| `--duration` | `-d` | int | `5` | No | Duration in seconds |
| `--negative` | - | string | - | No | Negative prompt (max 500 chars) |
| `--audio` | - | bool | `false` | No | Enable auto audio generation (wan2.6 i2v-flash only) |
| `--no-audio` | - | bool | `false` | No | Disable audio for r2v-flash (r2v has audio on by default) |
| `--audio-url` | - | string | - | No | Custom audio URL (wav/mp3, 3-30s, ≤15MB) |
| `--shot-type` | - | string | - | No | Shot type: single, multi (wan2.6 only) |
| `--prompt-extend` | - | bool | `true` | No | Enable prompt smart rewriting |
| `--watermark` | - | bool | `false` | No | Add "AI generated" watermark |
| `--seed` | - | int | - | No | Random seed [0, 2147483647] |

### Model Auto-Selection

| Input Type | Default Model | Alternatives |
|------------|---------------|--------------|
| Text only | `wan2.6-t2v` | wan2.5-t2v-preview, wan2.2-t2v-plus, wanx2.1-t2v-turbo, wanx2.1-t2v-plus |
| Image (--image) | `wan2.6-i2v-flash` | wan2.6-i2v, wan2.5-i2v-preview, wan2.2-i2v-flash, wan2.2-i2v-plus, wanx2.1-i2v-turbo, wanx2.1-i2v-plus |
| Reference (--ref) | `wan2.6-r2v-flash` | wan2.6-r2v |
| Keyframe (--first-frame) | `wan2.2-kf2v-flash` | wanx2.1-kf2v-plus |

### Duration Rules

| Model Series | Supported Duration |
|-------------|-------------------|
| wan2.6 (t2v) | 2-15 seconds (integer) |
| wan2.6 (r2v) | 2-10 seconds (integer) |
| wan2.5 | 5 or 10 seconds |
| wan2.2 | 5 seconds |
| wanx2.1-turbo | 3, 4, or 5 seconds |
| wanx2.1-plus | 5 seconds |
| kf2v (all) | 5 seconds (fixed) |

### Resolution Rules

| Model Series | Supported Resolutions |
|-------------|----------------------|
| wan2.6 | 720P, 1080P |
| wan2.5 | 480P, 720P, 1080P |
| wan2.2-t2v-plus | 480P, 1080P |
| wan2.2-i2v-flash | 480P, 720P, 1080P |
| wan2.2-i2v-plus | 480P, 1080P |
| wanx2.1-turbo | 480P, 720P |
| wanx2.1-plus | 720P |
| kf2v-flash | 480P, 720P, 1080P |
| kf2v-plus | 720P |

### Image Input (--image, --first-frame, --last-frame)

- Formats: JPEG, PNG, BMP, WEBP
- Dimensions: 360-2000 px per side
- Max file size: 10MB
- Supports: local file path, public URL, Base64

### Reference Input (--ref)

Reference files can be images or videos. Use `character1`, `character2`, etc. in prompt to reference them by order.

**Images:**
- Formats: JPEG, PNG, BMP, WEBP
- Dimensions: 240-5000 px per side
- Max file size: 10MB
- No transparent channel support

**Videos:**
- Formats: MP4, MOV
- Duration: 1-30 seconds
- Max file size: 100MB

**Limits:**
- Images: 0-5
- Videos: 0-3
- Total: ≤5 files

### Audio Requirements

- `--audio`: auto-generate audio (only wan2.6-i2v-flash)
- `--no-audio`: disable audio for r2v-flash (r2v has audio enabled by default)
- `--audio-url`: custom audio file URL (wan2.5+ t2v/i2v)
  - Formats: WAV, MP3
  - Duration: 3-30 seconds
  - Max size: 15MB

### Mutual Exclusivity

- `--image`, `--ref`, and `--first-frame` cannot be used together (pick one input mode)
- `--last-frame` requires `--first-frame`
- `--ratio` is ignored for i2v/kf2v (ratio comes from input image)
- `--audio` only works with wan2.6-i2v-flash model
- `--no-audio` only works with wan2.6-r2v-flash model
- `--audio-url` only works with wan2.5+ t2v/i2v models (not r2v)
- `--shot-type` only works with wan2.6 models
- `--prompt-extend` is not available for r2v models

### Output

```json
{
  "success": true,
  "task_id": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  "status": "pending"
}
```

---

## dashscope video status

Query the status of a video generation task.

### Usage

```bash
rawgenai dashscope video status <task_id>
rawgenai dashscope video status <task_id> -v  # Show full output including URLs
```

### Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `--verbose` | `-v` | bool | `false` | No | Show full output including video URL |

### Output

**Pending/Running:**
```json
{
  "success": true,
  "task_id": "xxx",
  "status": "running"
}
```

**Succeeded:**
```json
{
  "success": true,
  "task_id": "xxx",
  "status": "succeeded",
  "duration": 10,
  "resolution": 720
}
```

**Succeeded (with --verbose):**
```json
{
  "success": true,
  "task_id": "xxx",
  "status": "succeeded",
  "duration": 10,
  "resolution": 720,
  "video_url": "https://...",
  "orig_prompt": "original prompt",
  "actual_prompt": "rewritten prompt"
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

| Status | API Status | Description |
|--------|-----------|-------------|
| `pending` | PENDING | Task is waiting in queue |
| `running` | RUNNING | Video is being generated |
| `succeeded` | SUCCEEDED | Video generation completed |
| `failed` | FAILED | Generation failed |

---

## dashscope video download

Download a completed video.

### Usage

```bash
rawgenai dashscope video download <task_id> -o output.mp4
```

### Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `--output` | `-o` | string | - | Yes | Output file path (.mp4) |

### Output

```json
{
  "success": true,
  "task_id": "xxx",
  "file": "/absolute/path/to/output.mp4"
}
```

### Notes

- Video URL is only valid for **24 hours** after task completion
- Download will fail if the URL has expired

---

## Environment Variables

| Variable | Description |
|----------|-------------|
| `DASHSCOPE_API_KEY` | DashScope API key (required) |
| `DASHSCOPE_BASE_URL` | Custom base URL (optional, default: Beijing region) |

### API Regions

| Region | Base URL |
|--------|----------|
| Beijing (default) | `https://dashscope.aliyuncs.com/api/v1` |
| Singapore | `https://dashscope-intl.aliyuncs.com/api/v1` |
| Virginia | `https://dashscope-us.aliyuncs.com/api/v1` |

To switch regions:

```bash
# Via config (persistent)
rawgenai config set dashscope_base_url "https://dashscope-intl.aliyuncs.com/api/v1"

# Or via environment variable
export DASHSCOPE_BASE_URL="https://dashscope-intl.aliyuncs.com/api/v1"
```

> API Keys are region-specific and cannot be used across regions.

---

## Errors

### CLI Errors

| Code | Description |
|------|-------------|
| `missing_api_key` | DASHSCOPE_API_KEY not set |
| `missing_prompt` | No prompt provided (required for t2v) |
| `missing_task_id` | Task ID not provided |
| `missing_output` | Output file path required |
| `file_not_found` | Input file not found |
| `image_not_found` | Image file not found |
| `image_read_error` | Cannot read image file |
| `first_frame_not_found` | First frame image not found |
| `last_frame_not_found` | Last frame image not found |
| `last_frame_requires_first` | --last-frame requires --first-frame |
| `conflicting_input_flags` | Cannot use --image, --ref, and --first-frame together |
| `missing_ref` | --ref requires at least one reference file |
| `too_many_refs` | Too many reference files (max 5 total) |
| `too_many_ref_videos` | Too many reference videos (max 3) |
| `ref_not_found` | Reference file not found |
| `ref_read_error` | Cannot read reference file |
| `invalid_model` | Invalid model name |
| `invalid_resolution` | Resolution not supported by this model |
| `invalid_duration` | Duration not supported by this model |
| `invalid_ratio` | Invalid aspect ratio (use 16:9 or 9:16) |
| `invalid_format` | Invalid output format |
| `incompatible_audio` | --audio only supported by wan2.6-i2v-flash |
| `incompatible_audio_url` | --audio-url only supported by wan2.5+ models |
| `incompatible_shot_type` | --shot-type only supported by wan2.6 models |
| `incompatible_no_audio` | --no-audio only supported by wan2.6-r2v-flash |
| `output_write_error` | Cannot write output file |

### API Errors

| Code | Description |
|------|-------------|
| `invalid_api_key` | API key is invalid or region mismatch |
| `invalid_request` | Invalid request parameters |
| `content_policy` | Content violates safety policy |
| `task_not_found` | Task ID not found or expired (24h) |
| `rate_limit` | Too many requests (default QPS: 20) |
| `server_error` | DashScope server error |

### Download Errors

| Code | Description |
|------|-------------|
| `video_not_ready` | Video generation not completed |
| `video_failed` | Video generation failed |
| `no_video` | No video URL in response |
| `url_expired` | Video URL expired (24h limit) |
| `download_error` | Cannot download video |
| `connection_error` | Network connection failed |
| `timeout` | Request timed out |

---

## Internal: Resolution to Size Mapping (t2v, r2v)

The t2v and r2v APIs use `size` parameter (e.g., `1280*720`) while i2v/kf2v use `resolution` (e.g., `720P`). The CLI unifies this to `--resolution` and converts internally for t2v/r2v:

| Resolution | Ratio 16:9 | Ratio 9:16 |
|-----------|------------|------------|
| 480P | `832*480` | `480*832` |
| 720P | `1280*720` | `720*1280` |
| 1080P | `1920*1080` | `1080*1920` |
