# rawgenai google image

Generate and edit images using Google Gemini Nano Banana models.

## Usage

```bash
rawgenai google image <prompt> [flags]
rawgenai google image --file <prompt.txt> [flags]
cat prompt.txt | rawgenai google image [flags]
```

## Examples

```bash
# Text to image
rawgenai google image "A cute cat wearing a hat" -o cat.png

# High resolution with Pro model (up to 4K)
rawgenai google image "Mountain landscape at sunset" -o landscape.png --model pro --size 4K

# Specific aspect ratio
rawgenai google image "A portrait photo" -o portrait.png --aspect 9:16

# Landscape aspect ratio
rawgenai google image "Ocean panorama" -o ocean.png --aspect 21:9

# Image editing (with reference image)
rawgenai google image "Make it a watercolor painting" --image photo.png -o watercolor.png

# Multiple reference images (Pro model, up to 14)
rawgenai google image "Combine these into one scene" --image img1.png --image img2.png --image img3.png -o combined.png --model pro

# From file
rawgenai google image --file prompt.txt -o output.png

# From stdin
echo "A forest at dawn" | rawgenai google image -o forest.png

# Pro model with Google Search grounding
rawgenai google image "Current weather in Tokyo as an infographic" -o weather.png --model pro --search
```

## Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `--output` | `-o` | string | - | Yes | Output file path (.png) |
| `--image` | `-i` | string[] | - | No | Reference image(s), can be repeated |
| `--file` | - | string | - | No | Input prompt file |
| `--model` | `-m` | string | `flash` | No | Model: flash, pro |
| `--aspect` | `-a` | string | `1:1` | No | Aspect ratio |
| `--size` | `-s` | string | `1K` | No | Image size (Pro only): 1K, 2K, 4K |
| `--search` | - | bool | `false` | No | Enable Google Search grounding (Pro only) |

## Models

| Model | Model ID | Description |
|-------|----------|-------------|
| `flash` | `gemini-2.5-flash-image` | Fast, efficient (default), ~1024px |
| `pro` | `gemini-3-pro-image-preview` | High quality, up to 4K, advanced features |

### Model Comparison

| Feature | Flash (Nano Banana) | Pro (Nano Banana Pro) |
|---------|---------------------|----------------------|
| Resolution | Fixed ~1024px | 1K, 2K, 4K |
| Max Input Images | 3 | 14 (5 high-fidelity) |
| Google Search | No | Yes |
| Thinking Mode | No | Yes (auto) |
| Text Rendering | Basic | Advanced |
| Speed | Faster | Slower |

## Aspect Ratios

| Value | Description |
|-------|-------------|
| `1:1` | Square (default) |
| `2:3` | Portrait |
| `3:2` | Landscape |
| `3:4` | Portrait |
| `4:3` | Landscape |
| `4:5` | Portrait |
| `5:4` | Landscape |
| `9:16` | Vertical (mobile) |
| `16:9` | Horizontal (widescreen) |
| `21:9` | Ultra-wide |

## Image Size (Pro Model Only)

| Value | Description |
|-------|-------------|
| `1K` | ~1024px (default) |
| `2K` | ~2048px |
| `4K` | ~4096px |

**Note:** Must use uppercase `K`. Lowercase will be rejected.

## Resolution Reference

### Flash Model (gemini-2.5-flash-image)

| Aspect | Resolution |
|--------|------------|
| 1:1 | 1024x1024 |
| 2:3 | 832x1248 |
| 3:2 | 1248x832 |
| 9:16 | 768x1344 |
| 16:9 | 1344x768 |
| 21:9 | 1536x672 |

### Pro Model (gemini-3-pro-image-preview)

| Aspect | 1K | 2K | 4K |
|--------|----|----|----|
| 1:1 | 1024x1024 | 2048x2048 | 4096x4096 |
| 16:9 | 1376x768 | 2752x1536 | 5504x3072 |
| 9:16 | 768x1376 | 1536x2752 | 3072x5504 |

## Input Images

Reference images for editing or composition:

- **Flash model:** Best with up to 3 images
- **Pro model:** Up to 14 images total
  - Up to 6 object images (high-fidelity)
  - Up to 5 human images (character consistency)

## Output Format

Output is always PNG format.

## Output

```json
{
  "success": true,
  "file": "/path/to/output.png",
  "model": "gemini-2.5-flash-image",
  "aspect": "16:9"
}
```

Pro model with size:

```json
{
  "success": true,
  "file": "/path/to/output.png",
  "model": "gemini-3-pro-image-preview",
  "aspect": "16:9",
  "size": "2K"
}
```

## Errors

```json
{
  "success": false,
  "error": {
    "code": "invalid_aspect",
    "message": "Aspect ratio '1:2' is not valid"
  }
}
```

### CLI Errors (before API call)

| Code | Description |
|------|-------------|
| `missing_api_key` | GEMINI_API_KEY not set |
| `missing_prompt` | No prompt provided |
| `missing_output` | --output flag not provided |
| `file_not_found` | Input file does not exist |
| `image_not_found` | Reference image file does not exist |
| `unsupported_format` | Output file extension not .png |
| `invalid_model` | Model not flash or pro |
| `invalid_aspect` | Aspect ratio not valid |
| `invalid_size` | Size not 1K, 2K, or 4K |
| `size_requires_pro` | --size flag requires pro model |
| `search_requires_pro` | --search flag requires pro model |
| `too_many_images` | Too many reference images for model |
| `output_write_error` | Cannot write to output file |

### Gemini API Errors

| HTTP | Code | Description |
|------|------|-------------|
| 400 | `invalid_request` | Invalid request parameters |
| 400 | `content_policy` | Content violates safety policy |
| 401 | `invalid_api_key` | API key is invalid or revoked |
| 403 | `permission_denied` | API key lacks required permissions |
| 429 | `rate_limit` | Too many requests |
| 429 | `quota_exceeded` | Quota exhausted |
| 500 | `server_error` | Gemini server error |
| 503 | `server_overloaded` | Gemini server overloaded |

### Network Errors

| Code | Description |
|------|-------------|
| `connection_error` | Cannot connect to Gemini API |
| `timeout` | Request timed out |

## Limitations

- Best performance languages: EN, zh-CN, ja-JP, ko-KR, de-DE, fr-FR, es-MX, pt-BR, it-IT, ru-RU, ar-EG, hi-IN, id-ID, vi-VN
- Audio/video inputs not supported
- Model may not always generate exact requested number of images
- All generated images include SynthID watermark
- For best text rendering, generate text first then request image with that text

## Prompting Tips

- **Describe the scene**, don't just list keywords
- Be specific: "ornate elven plate armor" vs "fantasy armor"
- Provide context and intent
- Use photography terms for realistic images (lens, lighting, angle)
- Iterate and refine with follow-up prompts
