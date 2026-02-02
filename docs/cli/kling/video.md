# Kling Video CLI Design

## Overview

Kling video CLI uses the **Omni-Video O1** model (`kling-video-o1`), the latest unified multimodal video generation interface.

**API Endpoint:** `/v1/videos/omni-video`

**Capabilities:**
- Text-to-video
- Image-to-video (first frame, first+last frame)
- Reference images (character, style, scene)
- Reference video (style, camera movement)
- Video editing (add/remove/replace elements)

---

## Commands

```
kling video create [prompt]              # Create video (O1 model)
kling video create-from-text [prompt]    # Create video from text (legacy models)
kling video create-from-image [prompt]   # Create video from image (legacy models)
kling video create-motion-control [prompt] # Transfer motion from video to image
kling video create-avatar [prompt]       # Create digital avatar with lip sync
kling video status <task_id>             # Query task status
kling video download <task_id>           # Download completed video
kling video list                         # List video generation tasks
kling video extend <video_id>            # Extend an existing video
kling video add-sound <video_id>         # Add sound effects/BGM to video
kling video element create <name>        # Create a custom element
kling video element list                 # List elements (custom/official)
kling video element delete <id>          # Delete a custom element
```

## Documentation

| Document | Description |
|----------|-------------|
| [video-create.md](video-create.md) | `create` command (O1 model) |
| [video-legacy.md](video-legacy.md) | `create-from-text`, `create-from-image`, `create-motion-control`, `create-avatar` |
| [video-operations.md](video-operations.md) | `status`, `download`, `list`, `extend`, `add-sound` |
| [video-element.md](video-element.md) | Element management commands |
| [video-examples.md](video-examples.md) | Usage examples and workflows |
| [voice.md](voice.md) | Voice cloning commands (`kling voice`) |

---

## Authentication

Kling uses JWT authentication with Access Key + Secret Key:

| Environment Variable | Description |
|---------------------|-------------|
| `KLING_ACCESS_KEY` | Access key from Kling AI console |
| `KLING_SECRET_KEY` | Secret key from Kling AI console |

JWT token is generated dynamically with 30-minute expiry.

### API Region

Kling provides separate API endpoints for China and International regions:

| Region | Base URL |
|--------|----------|
| China (default) | `https://api-beijing.klingai.com` |
| International | `https://api-singapore.klingai.com` |

To switch regions, set `KLING_BASE_URL` environment variable or use config:

```bash
# Set via config (persistent)
rawgenai config set kling_base_url "https://api-singapore.klingai.com"

# Or via environment variable
export KLING_BASE_URL="https://api-singapore.klingai.com"
```

**Priority:** Environment variable > Config file > Default (China)

---

## Error Codes

### CLI Errors

| Code | Description |
|------|-------------|
| `missing_prompt` | No prompt provided |
| `missing_api_key` | KLING_ACCESS_KEY or KLING_SECRET_KEY not set |
| `missing_task_id` | Task ID argument required |
| `missing_video_id` | Video ID argument required |
| `missing_output` | Output file path required |
| `invalid_mode` | Invalid mode (use std or pro) |
| `invalid_duration` | Invalid duration for this mode |
| `invalid_ratio` | Invalid aspect ratio |
| `invalid_format` | Unsupported output format |
| `invalid_model` | Invalid model name |
| `invalid_camera_control` | Invalid camera control JSON |
| `invalid_dynamic_mask` | Invalid dynamic mask JSON |
| `frame_not_found` | Frame image file not found |
| `frame_read_error` | Cannot read frame image file |
| `image_not_found` | Image file not found |
| `ref_image_not_found` | Reference image file not found |
| `mask_not_found` | Mask image file not found |
| `mask_read_error` | Cannot read mask image file |
| `last_frame_requires_first` | --last-frame requires --first-frame |
| `conflicting_video_flags` | Cannot use --ref-video and --base-video together |
| `too_many_images` | Too many images (max 7 without video, 4 with video) |
| `missing_name` | Element name is required |
| `invalid_name` | Element name exceeds 20 characters |
| `missing_description` | Element description is required |
| `invalid_description` | Element description exceeds 100 characters |
| `missing_frontal` | Frontal image is required |
| `frontal_not_found` | Frontal image file not found |
| `invalid_ref_count` | Must provide 1-3 reference images |
| `ref_not_found` | Reference image file not found |
| `invalid_tag` | Invalid element tag |
| `missing_element_id` | Element ID is required |
| `invalid_type` | Invalid element type (use custom or official) |
| `missing_image` | Image is required |
| `missing_audio` | Audio file or audio ID is required |
| `conflicting_audio` | Cannot use both --audio and --audio-id |
| `audio_not_found` | Audio file not found |
| `audio_read_error` | Cannot read audio file |
| `incompatible_camera_control` | Camera control not supported by this model/mode/duration |
| `incompatible_sound` | Sound generation not supported by this model |
| `incompatible_voice` | Voice control not supported by this model |
| `incompatible_last_frame` | First+last frame not supported by this model/mode/duration |
| `incompatible_motion_brush` | Motion brush not supported by this model/mode/duration |

### API Errors

| Code | HTTP | Description |
|------|------|-------------|
| `invalid_api_key` | 401 | Invalid or expired API key |
| `permission_denied` | 403 | API key lacks permissions |
| `task_not_found` | 404 | Task ID not found |
| `invalid_request` | 400 | Invalid request parameters |
| `content_policy` | 400 | Content violates safety policy |
| `rate_limit` | 429 | Too many requests |
| `quota_exceeded` | 429 | Account quota exhausted |
| `server_error` | 500 | Kling server error |
| `server_overloaded` | 503 | Service temporarily unavailable |

### Network Errors

| Code | Description |
|------|-------------|
| `timeout` | Request timed out |
| `connection_error` | Cannot connect to Kling API |
