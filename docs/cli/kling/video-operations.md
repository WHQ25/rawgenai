# Kling Video Operations

Commands for managing video tasks: status, download, list, extend, add-sound.

---

## `kling video status`

Query the status of a video generation task.

### Usage

```bash
kling video status <task_id>
kling video status <task_id> -v  # Show full output including URLs
```

### Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `--type` | `-t` | string | `create` | No | Task type: create, text2video, image2video, extend, add-sound |
| `--verbose` | `-v` | bool | `false` | No | Show full output including URLs |

### Output

**Processing:**
```json
{
  "success": true,
  "task_id": "xxx",
  "status": "processing"
}
```

**Succeeded:**
```json
{
  "success": true,
  "task_id": "xxx",
  "status": "succeed",
  "video_id": "xxx",
  "duration": "5"
}
```

**Succeeded (with --verbose):**
```json
{
  "success": true,
  "task_id": "xxx",
  "status": "succeed",
  "video_id": "xxx",
  "duration": "5",
  "video_url": "https://...",
  "watermark_url": "https://..."
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

### Task Status Values

| Status | Description |
|--------|-------------|
| `submitted` | Task created, waiting to process |
| `processing` | Video is being generated |
| `succeed` | Video generation completed |
| `failed` | Generation failed |

---

## `kling video download`

Download a completed video or audio to local file.

### Usage

```bash
# Download video
kling video download <task_id> -o output.mp4

# Download audio (add-sound tasks only)
kling video download <task_id> --type add-sound --format mp3 -o output.mp3
kling video download <task_id> --type add-sound --format wav -o output.wav
```

### Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `--output` | `-o` | string | | Yes | Output file path |
| `--type` | `-t` | string | `create` | No | Task type: create, text2video, image2video, extend, add-sound |
| `--format` | | string | `video` | No | Download format: video, mp3, wav (mp3/wav only for add-sound) |
| `--watermark` | | bool | `false` | No | Download watermarked version |

### Output

```json
{
  "success": true,
  "task_id": "xxx",
  "file": "/absolute/path/to/output.mp4"
}
```

---

## `kling video list`

List video generation tasks.

### Usage

```bash
kling video list
kling video list --limit 50
```

### Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `--limit` | `-l` | int | `30` | No | Maximum tasks to return (1-500) |
| `--page` | `-p` | int | `1` | No | Page number |
| `--type` | `-t` | string | `create` | No | Task type: create, text2video, image2video, extend, add-sound |

### Output

```json
{
  "success": true,
  "tasks": [
    {
      "task_id": "xxx",
      "status": "succeed",
      "created_at": 1722769557708
    }
  ],
  "count": 1
}
```

---

## `kling video extend`

Extend an existing video by 4-5 seconds.

### Usage

```bash
kling video extend <video_id>
kling video extend <video_id> --prompt "Continue the action"
```

### Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `--prompt` | | string | | No | Prompt for the extended portion |
| `--negative` | | string | | No | Negative prompt |
| `--cfg-scale` | | float | `0.5` | No | Prompt adherence (0-1) |
| `--watermark` | | bool | `false` | No | Include watermark |

### Notes

- Extends video by 4-5 seconds per call
- Can be called repeatedly (max total duration: 3 minutes)
- Original video must be within 30 days
- Uses same model/mode as original video
- **Only works with legacy models** (not O1)

### Output

```json
{
  "success": true,
  "task_id": "xxx",
  "status": "submitted"
}
```

### Checking Status & Download

```bash
kling video status <task_id> --type extend
kling video download <task_id> --type extend -o extended.mp4
```

---

## `kling video add-sound`

Generate and add sound effects (foley) or background music to a video.

### Usage

```bash
# Using video ID (from Kling-generated video)
kling video add-sound <video_id>

# Using video URL (external video)
kling video add-sound --url "https://example.com/video.mp4"

# With sound effect prompt
kling video add-sound <video_id> --sound "footsteps, birds chirping"

# With background music prompt
kling video add-sound <video_id> --bgm "upbeat jazz music"

# Combined
kling video add-sound <video_id> --sound "ocean waves" --bgm "relaxing ambient"
```

### Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `--url` | | string | | No | Video URL (alternative to video_id) |
| `--sound` | `-s` | string | | No | Sound effect prompt (max 200 chars) |
| `--bgm` | `-b` | string | | No | Background music prompt (max 200 chars) |
| `--asmr` | | bool | `false` | No | Enable ASMR mode (enhanced detail) |

### Video Requirements

- Duration: 3-20 seconds
- Format: MP4, MOV
- Size: ≤100MB

### Output

```json
{
  "success": true,
  "task_id": "xxx",
  "status": "submitted"
}
```

### Checking Status & Download

```bash
# Check status (with verbose to see audio URLs)
kling video status <task_id> --type add-sound -v

# Download video with sound
kling video download <task_id> --type add-sound -o with_sound.mp4

# Download audio only
kling video download <task_id> --type add-sound --format mp3 -o audio.mp3
kling video download <task_id> --type add-sound --format wav -o audio.wav
```

Status output (with --verbose):
```json
{
  "success": true,
  "task_id": "xxx",
  "status": "succeed",
  "video_url": "https://...",
  "audio_mp3_url": "https://...",
  "audio_wav_url": "https://..."
}
```
