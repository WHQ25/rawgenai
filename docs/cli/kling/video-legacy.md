# Kling Video Legacy Commands

Commands for legacy Kling video models and motion control.

---

## `kling video create-from-text`

Create video from text prompt using legacy models.

### Usage

```bash
kling video create-from-text "A cat playing piano"
kling video create-from-text "A sunset" --model kling-v2-6 --sound
```

### Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `--model` | `-m` | string | `kling-v1` | No | Model name |
| `--mode` | | string | `std` | No | Generation mode: std, pro |
| `--duration` | `-d` | string | `5` | No | Video duration: 5, 10 |
| `--ratio` | `-r` | string | `16:9` | No | Aspect ratio: 16:9, 9:16, 1:1 |
| `--negative` | | string | | No | Negative prompt |
| `--cfg-scale` | | float | `0.5` | No | Prompt adherence (0-1), not supported by v2.x |
| `--camera-control` | | string | | No | Camera control JSON |
| `--sound` | | bool | `false` | No | Generate sound (v2.6+ only) |
| `--watermark` | | bool | `false` | No | Include watermark |
| `--prompt-file` | `-f` | string | | No | Read prompt from file |

### Available Models

| Model | Description |
|-------|-------------|
| `kling-v1` | Original Kling model |
| `kling-v1-6` | Kling v1.6 |
| `kling-v2-master` | Kling v2 Master |
| `kling-v2-1-master` | Kling v2.1 Master |
| `kling-v2-5-turbo` | Kling v2.5 Turbo (faster) |
| `kling-v2-6` | Kling v2.6 (supports sound) |

### Compatibility Notes

| Feature | Supported Models/Modes |
|---------|----------------------|
| `--camera-control` | kling-v1 std 5s only |
| `--sound` | kling-v2-6 only |

### Camera Control

Control camera movement with JSON:

```bash
# Simple camera control
--camera-control '{"type":"simple","config":{"horizontal":5,"vertical":0,"pan":0,"tilt":0,"roll":0,"zoom":0}}'

# Preset camera movement
--camera-control '{"type":"preset","config":"zoom_in"}'
```

**Simple config options:**
- `horizontal`: -10 to 10 (left/right)
- `vertical`: -10 to 10 (up/down)
- `pan`: -10 to 10
- `tilt`: -10 to 10
- `roll`: -10 to 10
- `zoom`: -10 to 10

**Preset options:** `zoom_in`, `zoom_out`, `pan_left`, `pan_right`, `tilt_up`, `tilt_down`

### Output

```json
{
  "success": true,
  "task_id": "xxx",
  "status": "submitted"
}
```

---

## `kling video create-from-image`

Create video from image using legacy models.

### Usage

```bash
# Basic usage
kling video create-from-image -i first.png

# With prompt
kling video create-from-image "The character walks forward" -i first.png

# First + last frame
kling video create-from-image -i start.png --last-frame end.png

# With motion control
kling video create-from-image -i photo.png --static-mask mask.png
```

### Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `--first-frame` | `-i` | string | | Yes | First frame image |
| `--last-frame` | | string | | No | Last frame image |
| `--model` | `-m` | string | `kling-v1` | No | Model name |
| `--mode` | | string | `std` | No | Generation mode: std, pro |
| `--duration` | `-d` | string | `5` | No | Video duration: 5, 10 |
| `--negative` | | string | | No | Negative prompt |
| `--cfg-scale` | | float | `0.5` | No | Prompt adherence (0-1), not supported by v2.x |
| `--camera-control` | | string | | No | Camera control JSON |
| `--static-mask` | | string | | No | Static brush mask image |
| `--dynamic-mask` | | string | | No | Dynamic mask JSON |
| `--voice` | | string[] | | No | Voice ID(s) for v2.6+ |
| `--sound` | | bool | `false` | No | Generate sound (v2.6+ only) |
| `--watermark` | | bool | `false` | No | Include watermark |
| `--prompt-file` | `-f` | string | | No | Read prompt from file |

### Available Models

| Model | Description |
|-------|-------------|
| `kling-v1` | Original Kling model |
| `kling-v1-5` | Kling v1.5 |
| `kling-v1-6` | Kling v1.6 |
| `kling-v2-master` | Kling v2 Master |
| `kling-v2-1` | Kling v2.1 |
| `kling-v2-1-master` | Kling v2.1 Master |
| `kling-v2-5-turbo` | Kling v2.5 Turbo |
| `kling-v2-6` | Kling v2.6 (supports sound, voice) |

### Compatibility Notes

| Feature | Supported Models/Modes |
|---------|----------------------|
| `--last-frame` | kling-v1: 5s only; kling-v1-5/v1-6/v2-1/v2-5-turbo/v2-6: pro mode only; kling-v2-master/v2-1-master: not supported |
| `--static-mask` | kling-v1: 5s only; kling-v1-5: pro 5s only; other models: not supported |
| `--dynamic-mask` | kling-v1: 5s only; kling-v1-5: pro 5s only; other models: not supported |
| `--camera-control` | kling-v1-5 pro 5s only (simple type) |
| `--sound` | kling-v2-6 only |
| `--voice` | kling-v2-6 only |

### Static Mask

Keep parts of the image static during video generation:

```bash
# Local file
--static-mask mask.png

# URL
--static-mask "https://example.com/mask.png"
```

The mask should be a grayscale image where:
- White areas: Keep static (no movement)
- Black areas: Allow movement

### Dynamic Mask

Control movement trajectories for specific regions:

```bash
--dynamic-mask '[{"mask":"mask.png","trajectories":[[{"x":100,"y":100},{"x":200,"y":150},{"x":300,"y":100}]]}]'
```

**Structure:**
```json
[
  {
    "mask": "path/to/mask.png",
    "trajectories": [
      [
        {"x": 100, "y": 100},
        {"x": 200, "y": 150},
        {"x": 300, "y": 100}
      ]
    ]
  }
]
```

- `mask`: Image file (local path or URL) - white areas will follow the trajectory
- `trajectories`: Array of point arrays defining movement paths

### Voice (v2.6+)

Add voice to generated video:

```bash
--voice voice_id_1,voice_id_2
```

### Output

```json
{
  "success": true,
  "task_id": "xxx",
  "status": "submitted"
}
```

---

---

## `kling video create-motion-control`

Transfer motion from a reference video to a reference image.

### Usage

```bash
# Basic usage
kling video create-motion-control -i person.png -v dance.mp4

# With prompt
kling video create-motion-control "Add particle effects" -i person.png -v dance.mp4

# Character faces same direction as video
kling video create-motion-control -i person.png -v dance.mp4 -o video

# Without original sound
kling video create-motion-control -i person.png -v dance.mp4 --keep-sound=false
```

### Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `--image` | `-i` | string | | Yes | Reference image |
| `--video` | `-v` | string | | Yes | Reference video for motion |
| `--orientation` | `-o` | string | `image` | No | Character orientation: image, video |
| `--mode` | `-m` | string | `std` | No | Generation mode: std, pro |
| `--keep-sound` | | bool | `true` | No | Keep original video sound |
| `--watermark` | | bool | `false` | No | Include watermark |
| `--prompt-file` | `-f` | string | | No | Read prompt from file |

### Character Orientation

| Value | Description | Max Video Duration |
|-------|-------------|-------------------|
| `image` | Character faces same direction as in image | ≤10 seconds |
| `video` | Character faces same direction as in video | ≤30 seconds |

### Image Requirements

- Person should show clear upper body or full body with head
- Avoid extreme orientations (upside down, lying flat)
- Person should occupy reasonable portion of frame
- Supports realistic/stylized characters (human/humanoid animals)
- Formats: JPEG, PNG
- Size: ≤10MB
- Dimensions: 300px - 65536px
- Aspect ratio: 1:2.5 ~ 2.5:1

### Video Requirements

- Person should show clear upper body or full body
- Recommended: single person video (multi-person uses largest person)
- Recommended: real human motion
- Single continuous shot (no cuts or camera movement)
- Moderate motion speed works best
- Formats: MP4, MOV
- Size: ≤100MB
- Dimensions: 340px - 3850px
- Duration: 3-30 seconds (depends on orientation)

### Output

```json
{
  "success": true,
  "task_id": "xxx",
  "status": "submitted"
}
```

---

## `kling video create-avatar`

Create digital avatar video with lip sync from an image and audio.

### Usage

```bash
# Basic usage with audio file
kling video create-avatar -i avatar.png --audio speech.mp3

# With audio URL
kling video create-avatar -i avatar.png --audio "https://example.com/audio.mp3"

# With audio ID from TTS preview
kling video create-avatar -i avatar.png --audio-id audio_123

# With optional prompt
kling video create-avatar "Add subtle head movements" -i avatar.png --audio speech.mp3

# Pro mode
kling video create-avatar -i avatar.png --audio speech.mp3 -m pro
```

### Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `--image` | `-i` | string | | Yes | Avatar reference image (local file or URL) |
| `--audio` | `-a` | string | | No* | Audio file for lip sync (local file or URL) |
| `--audio-id` | | string | | No* | Audio ID from TTS preview |
| `--mode` | `-m` | string | `std` | No | Generation mode: std, pro |
| `--watermark` | | bool | `false` | No | Include watermark |
| `--prompt-file` | `-f` | string | | No | Read prompt from file |

*Either `--audio` or `--audio-id` is required, but not both.

### Image Requirements

- Portrait image with clear face visibility
- Frontal or slight angle view recommended
- Formats: JPEG, PNG
- Good lighting and resolution

### Audio Requirements

- Clear speech audio without excessive background noise
- Formats: MP3, WAV, M4A, AAC
- Moderate speaking speed works best

### Output

```json
{
  "success": true,
  "task_id": "xxx",
  "status": "submitted"
}
```

---

## Checking Status & Downloading

Use the standard commands with `--type` flag:

```bash
# Check status
kling video status <task_id> --type text2video
kling video status <task_id> --type image2video
kling video status <task_id> --type motion-control
kling video status <task_id> --type avatar

# Download
kling video download <task_id> --type text2video -o output.mp4
kling video download <task_id> --type image2video -o output.mp4
kling video download <task_id> --type motion-control -o output.mp4
kling video download <task_id> --type avatar -o output.mp4

# List tasks
kling video list --type text2video
kling video list --type image2video
kling video list --type motion-control
kling video list --type avatar
```
