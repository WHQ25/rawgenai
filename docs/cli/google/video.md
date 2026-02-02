# rawgenai google video

Generate video using Google Veo models.

## Commands

| Command | Description |
|---------|-------------|
| `create` | Create a video generation job |
| `extend` | Extend a previously generated video |
| `status` | Get video generation status |
| `download` | Download generated video |

---

## video create

Create a video generation job. Returns immediately with an operation ID for async workflow.

### Usage

```bash
rawgenai google video create <prompt> [flags]
rawgenai google video create --file <prompt.txt> [flags]
cat prompt.txt | rawgenai google video create [flags]
```

### Examples

```bash
# Basic generation
rawgenai google video create "A cat playing piano on stage"

# Specific aspect ratio and duration
rawgenai google video create "A sunset over the ocean" --aspect 16:9 --duration 8

# Portrait video (for mobile)
rawgenai google video create "A person walking in the rain" --aspect 9:16

# Higher resolution
rawgenai google video create "Aerial view of a mountain range" --resolution 1080p

# With first frame image (image-to-video)
rawgenai google video create "She turns around and smiles" --first-frame start.jpg

# With first and last frame (interpolation)
rawgenai google video create "Smooth transition between scenes" --first-frame start.jpg --last-frame end.jpg

# With reference images (veo-3.1 only, max 3)
rawgenai google video create "A robot walking in a city" --ref robot.jpg

# With multiple reference images
rawgenai google video create "Two characters talking" --ref char1.jpg --ref char2.jpg --ref background.jpg

# Negative prompt (what to avoid)
rawgenai google video create "A peaceful garden" --negative "people, animals, text"

# From file
rawgenai google video create --file prompt.txt

# From stdin
echo "A dog running in a park" | rawgenai google video create
```

### Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `--file` | - | string | - | No | Input prompt file |
| `--first-frame` | - | string | - | No | First frame image (JPEG/PNG) |
| `--last-frame` | - | string | - | No | Last frame image (JPEG/PNG), requires --first-frame |
| `--ref` | - | string[] | - | No | Reference image (repeatable, max 3) |
| `--model` | `-m` | string | `veo-3.1` | No | Model: veo-3.1, veo-3.1-fast |
| `--aspect` | `-a` | string | `16:9` | No | Aspect ratio: 16:9, 9:16 |
| `--resolution` | `-r` | string | `720p` | No | Resolution: 720p, 1080p, 4k |
| `--duration` | `-d` | int | `8` | No | Duration in seconds: 4, 6, 8 |
| `--negative` | - | string | - | No | Negative prompt (what to avoid) |
| `--seed` | - | int | - | No | Seed for reproducibility |

### Output

```json
{
  "success": true,
  "operation_id": "operations/generate-videos-abc123",
  "status": "running",
  "model": "veo-3.1-generate-preview",
  "aspect": "16:9",
  "resolution": "720p",
  "duration": 4
}
```

---

## video extend

Extend a previously generated Veo video by 7 seconds.

### Usage

```bash
rawgenai google video extend <operation_id> <prompt>
rawgenai google video extend <operation_id> --file <prompt.txt>
```

### Examples

```bash
# Extend a video with a new prompt
rawgenai google video extend "operations/generate-videos-abc123" "The camera follows the butterfly as it lands on a flower"

# Extend with a different model
rawgenai google video extend "operations/generate-videos-abc123" "Continue the scene" --model veo-3.1-fast

# With negative prompt
rawgenai google video extend "operations/generate-videos-abc123" "The sun sets" --negative "rain, clouds"
```

### Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `--file` | - | string | - | No | Input prompt file |
| `--model` | `-m` | string | `veo-3.1` | No | Model: veo-3.1, veo-3.1-fast |
| `--negative` | - | string | - | No | Negative prompt (what to avoid) |

### Output

```json
{
  "success": true,
  "operation_id": "operations/generate-videos-xyz789",
  "status": "running",
  "model": "veo-3.1-generate-preview"
}
```

### Limitations

- **720p only**: Video extension only supports 720p resolution
- **Source video**: Must be from a previous Veo generation (within 2 days)
- **Extension length**: Each extension adds ~7 seconds
- **Maximum extensions**: Up to 20 times (max total ~148 seconds)
- **Audio**: Voice cannot be effectively extended if not present in the last 1 second

---

## video status

Query the status of a video generation operation.

### Usage

```bash
rawgenai google video status <operation_id>
```

### Examples

```bash
# Check status
rawgenai google video status "operations/generate-videos-abc123"
```

### Output

```json
{
  "success": true,
  "operation_id": "operations/generate-videos-abc123",
  "status": "completed",
  "progress": 1.0
}
```

#### Status Values

| Status | Description |
|--------|-------------|
| `running` | Video is being generated |
| `completed` | Video is ready for download |
| `failed` | Generation failed (see error_message) |

#### Failed Status

```json
{
  "success": true,
  "operation_id": "operations/generate-videos-abc123",
  "status": "failed",
  "error_message": "Content policy violation"
}
```

---

## video download

Download video from a completed generation operation.

### Usage

```bash
rawgenai google video download <operation_id> -o <output_file>
```

### Examples

```bash
# Download video
rawgenai google video download "operations/generate-videos-abc123" -o my_video.mp4
```

### Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `--output` | `-o` | string | - | Yes | Output file path (.mp4) |

### Output

```json
{
  "success": true,
  "operation_id": "operations/generate-videos-abc123",
  "file": "/path/to/my_video.mp4"
}
```

---

## Workflow Example

Video generation is asynchronous. Use the following workflow:

```bash
# 1. Create generation job (returns immediately)
rawgenai google video create "A beautiful sunset over the ocean"
# Output: {"success":true,"operation_id":"operations/generate-videos-abc123","status":"running",...}

# 2. Check status periodically
rawgenai google video status "operations/generate-videos-abc123"
# Output: {"success":true,"operation_id":"operations/generate-videos-abc123","status":"running","progress":0.5}

# 3. When completed, download the video
rawgenai google video download "operations/generate-videos-abc123" -o sunset.mp4
# Output: {"success":true,"operation_id":"operations/generate-videos-abc123","file":"/path/to/sunset.mp4"}

# 4. (Optional) Extend the video
rawgenai google video extend "operations/generate-videos-abc123" "The camera pulls back to reveal the coastline"
# Output: {"success":true,"operation_id":"operations/generate-videos-xyz789","status":"running",...}

# 5. Check extend status and download
rawgenai google video status "operations/generate-videos-xyz789"
rawgenai google video download "operations/generate-videos-xyz789" -o sunset_extended.mp4
```

---

## Models

| Model | Model ID | Description |
|-------|----------|-------------|
| `veo-3.1` | `veo-3.1-generate-preview` | Latest Veo model (default) |
| `veo-3.1-fast` | `veo-3.1-fast-generate-preview` | Faster generation |

## Aspect Ratios

| Value | Description |
|-------|-------------|
| `16:9` | Landscape (default) |
| `9:16` | Portrait (mobile) |

## Resolution

| Value | Description |
|-------|-------------|
| `720p` | HD (default) |
| `1080p` | Full HD (8s duration only) |
| `4k` | 4K Ultra HD (8s duration only) |

## Duration

| Seconds | Description |
|---------|-------------|
| `4` | Short clip |
| `6` | Medium length |
| `8` | Longer clip (default, required for 1080p/4k) |

## First and Last Frame Images

Use `--first-frame` to provide a reference image as the first frame:

- Supports JPEG and PNG formats
- Image aspect ratio should match `--aspect` parameter
- Useful for image-to-video generation

Use `--last-frame` to provide a reference image as the last frame:

- Requires `--first-frame` to be set
- Supports JPEG and PNG formats
- Creates interpolation between first and last frames
- Useful for controlled transitions

## Reference Images

Reference images allow you to guide video generation with visual references for characters, objects, or scenes.

### Usage (`--ref`)

- Repeatable flag (use multiple times)
- Maximum 3 images per request
- Supports JPEG and PNG formats
- Useful for maintaining character/object consistency

```bash
# Single reference image
rawgenai google video create "A robot walking in a city" --ref robot.jpg

# Multiple reference images (person, clothing, accessory)
rawgenai google video create "A woman wearing a dress and sunglasses walks on the beach" \
  --ref woman.jpg --ref dress.jpg --ref sunglasses.jpg
```

### Important Notes

- **Cannot be used with frame images**: `--ref` and `--first-frame`/`--last-frame` are mutually exclusive
- Image quality matters: higher resolution reference images produce better results
- Keep reference images consistent with your prompt for best results

---

## Errors

```json
{
  "success": false,
  "error": {
    "code": "video_not_ready",
    "message": "Video is not ready for download, current status: running"
  }
}
```

### CLI Errors (before API call)

| Code | Description |
|------|-------------|
| `missing_api_key` | GEMINI_API_KEY not set |
| `missing_prompt` | No prompt provided |
| `missing_output` | --output flag not provided |
| `missing_operation_id` | operation_id argument not provided |
| `file_not_found` | Input file does not exist |
| `first_frame_not_found` | First frame file does not exist |
| `last_frame_not_found` | Last frame file does not exist |
| `last_frame_requires_first` | --last-frame requires --first-frame |
| `ref_not_found` | Reference image file does not exist |
| `too_many_refs` | More than 3 --ref images provided |
| `conflicting_image_options` | --ref used with --first-frame/--last-frame |
| `invalid_image_format` | Image format not supported |
| `invalid_format` | Output file extension not .mp4 |
| `invalid_model` | Model not valid |
| `invalid_aspect` | Aspect ratio not valid |
| `invalid_resolution` | Resolution not valid |
| `invalid_resolution_duration` | 1080p/4k only supports 8s duration |
| `invalid_duration` | Duration not valid (4, 6, 8) |
| `output_write_error` | Cannot write to output file |

### Gemini API Errors

| HTTP | Code | Description |
|------|------|-------------|
| 400 | `invalid_request` | Invalid request parameters |
| 400 | `content_policy` | Content violates safety policy |
| 401 | `invalid_api_key` | API key is invalid or revoked |
| 403 | `permission_denied` | API key lacks required permissions |
| 404 | `operation_not_found` | Operation ID does not exist |
| 429 | `rate_limit` | Too many requests |
| 429 | `quota_exceeded` | Quota exhausted |
| 500 | `server_error` | Gemini server error |
| 503 | `server_overloaded` | Gemini server overloaded |

### Download Errors

| Code | Description |
|------|-------------|
| `video_not_ready` | Video generation not completed yet |
| `video_failed` | Video generation failed |
| `no_video` | No video in response |
| `download_error` | Cannot download video from URL |

### Extend Errors

| Code | Description |
|------|-------------|
| `video_not_ready` | Source video generation not completed yet |
| `no_video` | No video found in source operation |
| `download_error` | Cannot download source video |

### Network Errors

| Code | Description |
|------|-------------|
| `connection_error` | Cannot connect to Gemini API |
| `timeout` | Request timed out |

---

## Prompt Tips

For best results, describe:
- **Shot type**: wide, close-up, tracking, aerial, POV
- **Subject**: what/who is in the video
- **Action**: what is happening, movement direction
- **Setting**: where it takes place, time of day
- **Lighting**: natural, dramatic, soft, neon
- **Style**: cinematic, documentary, animation

Example: "Cinematic tracking shot of a golden retriever running through autumn leaves in a sunlit forest, slow motion, warm color grading"

## Content Restrictions

- Content must comply with Google's usage policies
- Real people and public figures cannot be generated
- Copyrighted characters will be rejected
