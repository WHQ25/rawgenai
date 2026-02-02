# kling video create

Create a video generation task using Omni-Video O1 model.

## Usage

```bash
# Text-to-video
kling video create "A cat playing piano"

# Image-to-video (first frame)
kling video create "The cat starts dancing" --first-frame cat.png

# Image-to-video (first + last frame)
kling video create "Smooth transition" --first-frame start.png --last-frame end.png

# Reference images (use <<<image_N>>> in prompt)
kling video create "<<<image_1>>> walks in the city, style like <<<image_2>>>" \
  --ref-image character.png --ref-image style.png

# Reference video (style/camera)
kling video create "Same camera movement, a dog running" \
  --ref-video https://example.com/reference.mp4

# Video editing (base video, keeps original sound by default)
kling video create "Add a hat to <<<video_1>>> person" \
  --base-video https://example.com/input.mp4

# Combined: reference images + reference video
kling video create "<<<image_1>>> dancing, camera like <<<video_1>>>" \
  --ref-image dancer.png \
  --ref-video https://example.com/camera.mp4

# From prompt file
kling video create --prompt-file prompt.txt

# From stdin
echo "A beautiful sunset" | kling video create
```

## Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `--first-frame` | `-i` | string | | No | First frame image path |
| `--last-frame` | | string | | No | Last frame image path (requires --first-frame) |
| `--ref-image` | | string[] | | No | Reference image(s), use `<<<image_N>>>` in prompt |
| `--element` | | int64[] | | No | Element ID(s), use `<<<element_N>>>` in prompt |
| `--ref-video` | | string | | No | Reference video URL (style/camera reference) |
| `--base-video` | | string | | No | Base video URL for editing |
| `--ref-exclude-sound` | | bool | `false` | No | Exclude sound from ref/base video |
| `--prompt-file` | `-f` | string | | No | Read prompt from file |
| `--mode` | | string | `pro` | No | Generation mode: std, pro |
| `--duration` | `-d` | int | `5` | No | Video duration in seconds |
| `--ratio` | `-r` | string | `16:9` | No | Aspect ratio: 16:9, 9:16, 1:1 |
| `--watermark` | | bool | `false` | No | Include watermark |

## Video Input Types

| Flag | API `refer_type` | Use Case |
|------|------------------|----------|
| `--ref-video` | `feature` | Style reference, camera movement reference, generate next/previous shot |
| `--base-video` | `base` | Video editing (add/remove/replace elements in video) |

## Prompt Placeholders

O1 model supports referencing inputs via placeholders in prompt:

| Placeholder | Description |
|-------------|-------------|
| `<<<image_1>>>`, `<<<image_2>>>`, ... | Reference images (in order of --ref-image flags) |
| `<<<element_1>>>`, `<<<element_2>>>`, ... | Custom elements (in order of --element flags) |
| `<<<video_1>>>` | Reference or base video |

**Note:** First/last frame images are NOT referenced via placeholders - they're used automatically.

## Duration Rules

| Mode | Duration |
|------|----------|
| Text-to-video | 5 or 10 seconds |
| First frame only | 5 or 10 seconds |
| First + last frame | 3-10 seconds |
| Reference images/video | 3-10 seconds |
| Video editing (--base-video) | Auto (matches input video) |

## Aspect Ratio Rules

| Mode | --ratio |
|------|---------|
| Text-to-video | Required |
| Reference images (no first frame) | Required |
| Reference video (feature type) | Required |
| With first frame | Ignored (uses image ratio) |
| Video editing (--base-video) | Ignored (uses video ratio) |

## Input Limits

**Images:**
- Formats: JPEG, PNG
- Size: ≤10MB per image
- Dimensions: ≥300px
- Aspect ratio: 1:2.5 ~ 2.5:1
- Max count: 7 (without video), 4 (with video)
- With >2 images: last frame not supported

**Videos:**
- Formats: MP4, MOV
- Size: ≤200MB
- Duration: 3-10 seconds
- Resolution: 720px - 2160px
- Frame rate: 24-60fps
- Max count: 1

## Mutual Exclusivity

- `--ref-video` and `--base-video` cannot be used together
- `--first-frame` and `--ref-image` can be combined (first frame + reference images)

## Output

```json
{
  "success": true,
  "task_id": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  "status": "submitted"
}
```
