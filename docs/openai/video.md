# rawgenai openai video

Generate video using OpenAI Sora models.

## Usage

```bash
rawgenai openai video <prompt> [flags]
rawgenai openai video --file <prompt.txt> [flags]
cat prompt.txt | rawgenai openai video [flags]
```

## Examples

```bash
# Basic
rawgenai openai video "A cat playing piano on stage" -o cat.mp4

# High quality, longer duration
rawgenai openai video "A sunset over the ocean with waves" --model sora-2-pro --seconds 8 -o sunset.mp4

# Portrait video (for mobile)
rawgenai openai video "A person walking in the rain" --size 720x1280 -o walk.mp4

# Wide cinematic shot
rawgenai openai video "Aerial view of a mountain range" --size 1792x1024 -o mountain.mp4

# With first frame image (image-to-video)
rawgenai openai video "She turns around and smiles" --image first_frame.jpg -o scene.mp4

# From file
rawgenai openai video --file prompt.txt -o output.mp4

# From stdin
echo "A dog running in a park" | rawgenai openai video -o dog.mp4

# Don't wait (returns job ID)
rawgenai openai video "A rocket launch" -o rocket.mp4 --no-wait

# Custom timeout (5 minutes)
rawgenai openai video "Complex scene" -o scene.mp4 --timeout 300
```

## Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `--output` | `-o` | string | - | Yes | Output file path (must be .mp4) |
| `--file` | - | string | - | No | Input prompt file |
| `--image` | `-i` | string | - | No | First frame image (JPEG/PNG/WebP) |
| `--model` | `-m` | string | `sora-2` | No | Model name |
| `--size` | `-s` | string | `1280x720` | No | Video resolution |
| `--seconds` | - | int | `4` | No | Video duration (4, 8, 12) |
| `--no-wait` | - | bool | `false` | No | Return immediately with job ID |
| `--timeout` | - | int | `600` | No | Max wait time in seconds |

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

## Output

### Success (with wait)

```json
{
  "success": true,
  "file": "/path/to/output.mp4",
  "video_id": "video_abc123",
  "model": "sora-2",
  "size": "1280x720",
  "seconds": 4
}
```

### Success (with --no-wait)

```json
{
  "success": true,
  "video_id": "video_abc123",
  "status": "queued",
  "model": "sora-2",
  "size": "1280x720",
  "seconds": 4
}
```

## Errors

```json
{
  "success": false,
  "error": {
    "code": "generation_failed",
    "message": "Video generation failed: content policy violation"
  }
}
```

### CLI Errors (before API call)

| Code | Description |
|------|-------------|
| `missing_api_key` | OPENAI_API_KEY not set |
| `missing_prompt` | No prompt provided (no argument, empty file, empty stdin) |
| `missing_output` | --output flag not provided |
| `file_not_found` | Input file specified by --file does not exist |
| `image_not_found` | Image file specified by --image does not exist |
| `invalid_image_format` | Image format not supported (must be JPEG/PNG/WebP) |
| `invalid_format` | Output file extension is not .mp4 |
| `invalid_size` | Size value not in allowed list |
| `invalid_seconds` | Seconds value not in allowed list (4, 8, 12) |
| `output_write_error` | Cannot write to output file |

### OpenAI API Errors

| HTTP | Code | Description |
|------|------|-------------|
| 400 | `invalid_model` | Model does not exist |
| 400 | `content_policy` | Content violates usage policies |
| 400 | `invalid_request` | Other invalid request parameters |
| 401 | `invalid_api_key` | API key is invalid or revoked |
| 403 | `region_not_supported` | Region/country not supported |
| 429 | `rate_limit` | Too many requests |
| 429 | `quota_exceeded` | Quota/credits exhausted |
| 500 | `server_error` | OpenAI server error |
| 503 | `server_overloaded` | OpenAI server overloaded |

### Generation Errors

| Code | Description |
|------|-------------|
| `generation_failed` | Video generation failed (check message for details) |
| `timeout` | Generation did not complete within timeout |

### Network Errors

| Code | Description |
|------|-------------|
| `connection_error` | Cannot connect to OpenAI API |
| `timeout` | Request timed out |

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
