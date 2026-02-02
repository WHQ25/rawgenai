# rawgenai seed image

Generate and edit images using ByteDance Seedream models.

## Usage

```bash
rawgenai seed image <prompt> [flags]
rawgenai seed image --prompt-file <prompt.txt> [flags]
cat prompt.txt | rawgenai seed image [flags]
```

## Examples

```bash
# Text to image
rawgenai seed image "A cute cat wearing a hat" -o cat.jpg

# High resolution (4K)
rawgenai seed image "Mountain landscape at sunset" -o landscape.jpg --size 4K

# Custom resolution
rawgenai seed image "A portrait photo" -o portrait.jpg --size 2048x3072

# Image editing (with reference image)
rawgenai seed image "Make it a watercolor painting" --image photo.png -o watercolor.jpg

# Multiple reference images for fusion
rawgenai seed image "Replace the clothes in image 1 with the clothes in image 2" --image person.jpg --image clothes.jpg -o result.jpg

# Generate multiple images (comic/storyboard style)
rawgenai seed image "Generate 4 images showing the four seasons in the same garden" -o seasons.jpg -n 4

# With older model (faster, budget-friendly)
rawgenai seed image "A forest at dawn" -o forest.jpg --model 4.0

# From file
rawgenai seed image --prompt-file prompt.txt -o output.jpg

# From stdin
echo "A beautiful sunset" | rawgenai seed image -o sunset.jpg

# With watermark
rawgenai seed image "Product photo" -o product.jpg --watermark
```

## Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `--output` | `-o` | string | - | Yes | Output file path (.jpg/.jpeg) |
| `--image` | `-i` | string[] | - | No | Reference image(s), can be repeated (max 14) |
| `--prompt-file` | - | string | - | No | Input prompt file |
| `--model` | `-m` | string | `4.5` | No | Model: 4.5, 4.0 |
| `--size` | `-s` | string | `2K` | No | Image size: 2K, 4K, or WxH |
| `--count` | `-n` | int | `1` | No | Number of images (1-10) |
| `--watermark` | - | bool | `false` | No | Add watermark to output |

## Models

| Model | Model ID | Description |
|-------|----------|-------------|
| `4.5` | `doubao-seedream-4-5-251128` | Latest, best quality (default) |
| `4.0` | `doubao-seedream-4-0-250828` | Balanced quality and speed |

### Model Comparison

| Feature | Seedream 4.5 | Seedream 4.0 |
|---------|--------------|--------------|
| Quality | Highest | Good |
| Edit Consistency | Excellent | Good |
| Portrait Enhancement | Advanced | Basic |
| Text Rendering | Advanced | Basic |
| Multi-image Fusion | Enhanced | Standard |
| Speed | Slower | Faster |

## Image Size

### Preset Sizes

| Value | Description |
|-------|-------------|
| `2K` | ~2048px resolution (default) |
| `4K` | ~4096px resolution |

**Note:** When using preset sizes, describe the aspect ratio or shape in your prompt (e.g., "portrait photo", "wide panorama").

### Custom Sizes (WxH)

Specify exact dimensions in pixels (e.g., `2048x2048`, `1920x1080`).

**Constraints:**
- Width and height must be >= 14px
- Aspect ratio must be between 1:16 and 16:1
- Total pixels:
  - Seedream 4.5: 3,686,400 to 16,777,216 (approx. 1920x1920 to 4096x4096)
  - Seedream 4.0: 921,600 to 16,777,216 (approx. 960x960 to 4096x4096)

## Input Images

Reference images for editing or composition:

- Maximum 14 images
- Supported formats: JPEG, PNG, WebP, BMP, TIFF, GIF
- Max size: 10 MB per image
- Max total pixels: 36,000,000 per image

## Output Format

Output is always JPEG format.

## Output

### Single Image

```json
{
  "success": true,
  "file": "/path/to/output.jpg",
  "model": "doubao-seedream-4-5-251128",
  "size": "2K",
  "count": 1
}
```

### Multiple Images

```json
{
  "success": true,
  "files": [
    "/path/to/output_1.jpg",
    "/path/to/output_2.jpg",
    "/path/to/output_3.jpg",
    "/path/to/output_4.jpg"
  ],
  "model": "doubao-seedream-4-5-251128",
  "size": "2K",
  "count": 4
}
```

## Errors

```json
{
  "success": false,
  "error": {
    "code": "invalid_size",
    "message": "invalid size '3K', use '2K', '4K', or 'WxH' format"
  }
}
```

### CLI Errors (before API call)

| Code | Description |
|------|-------------|
| `missing_api_key` | ARK_API_KEY not set |
| `missing_prompt` | No prompt provided |
| `missing_output` | --output flag not provided |
| `file_not_found` | Input file does not exist |
| `image_not_found` | Reference image file does not exist |
| `image_read_error` | Cannot read reference image file |
| `unsupported_format` | Output file extension not .jpg/.jpeg |
| `invalid_model` | Model not 4.5 or 4.0 |
| `invalid_size` | Size not 2K, 4K, or valid WxH |
| `invalid_count` | Count not between 1 and 10 |
| `too_many_images` | More than 14 reference images |
| `decode_error` | Cannot decode image from response |
| `output_write_error` | Cannot write to output file |

### Ark API Errors

| HTTP | Code | Description |
|------|------|-------------|
| 400 | `invalid_request` | Invalid request parameters |
| 400 | `content_policy` | Content violates safety policy |
| 401 | `invalid_api_key` | API key is invalid or revoked |
| 403 | `permission_denied` | API key lacks required permissions |
| 429 | `rate_limit` | Too many requests (500 images/min limit) |
| 429 | `quota_exceeded` | Quota exhausted |
| 500 | `server_error` | Ark server error |
| 503 | `server_overloaded` | Ark server overloaded |

### Network Errors

| Code | Description |
|------|-------------|
| `connection_error` | Cannot connect to Ark API |
| `timeout` | Request timed out |

## Use Cases

### Text to Image
Generate images from text descriptions.

```bash
rawgenai seed image "A serene lake at sunset with mountains in the background" -o lake.jpg
```

### Image Editing
Edit existing images based on instructions.

```bash
rawgenai seed image "Change the dress color to red" --image original.jpg -o edited.jpg
```

### Multi-Image Fusion
Combine elements from multiple images.

```bash
rawgenai seed image "Put the person from image 1 in the background of image 2" \
  --image person.jpg --image background.jpg -o combined.jpg
```

### Sequential Image Generation
Generate a series of related images (comics, storyboards, brand visuals).

```bash
rawgenai seed image "Generate 4 panels showing a day in the life of a cat" -o comic.jpg -n 4
```

## Prompting Tips

- Use clear, natural language describing **subject + action + environment**
- Add style, color, lighting, composition details for aesthetic control
- Keep prompts under 300 Chinese characters or 600 English words
- For multi-image generation, describe the relationship between images
- Reference images by number (e.g., "image 1", "image 2") when using multiple inputs

## Limitations

- Best performance with: EN, zh-CN, ja-JP, ko-KR, de-DE, fr-FR, es-MX, pt-BR, it-IT, ru-RU, ar-EG, hi-IN, id-ID, vi-VN
- Generated image URLs expire after 24 hours
- Rate limit: 500 images per minute per model
- All generated images include invisible watermark (SynthID)
