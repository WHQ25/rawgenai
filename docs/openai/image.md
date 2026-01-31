# rawgenai openai image

Generate and edit images using OpenAI Responses API with GPT Image models.

## Usage

```bash
rawgenai openai image <prompt> [flags]
rawgenai openai image --file <prompt.txt> [flags]
cat prompt.txt | rawgenai openai image [flags]
```

## Examples

```bash
# Text to image
rawgenai openai image "A cute cat wearing a hat" -o cat.png

# Multi-turn conversation
rawgenai openai image "A landscape painting" -o v1.png
# Output: {"success":true,"file":"v1.png","response_id":"resp_abc123",...}

rawgenai openai image "Add a sunset to the sky" --continue resp_abc123 -o v2.png
# Output: {"success":true,"file":"v2.png","response_id":"resp_def456",...}

rawgenai openai image "Make it more dramatic" --continue resp_def456 -o v3.png

# High quality, specific size
rawgenai openai image "Mountain landscape at sunset" -o landscape.png --quality high --size 1536x1024

# With reference image
rawgenai openai image "Make it a watercolor painting" --image photo.png -o watercolor.png

# Multiple reference images
rawgenai openai image "Combine these into a gift basket" --image item1.png --image item2.png --image item3.png -o basket.png

# Inpainting with mask
rawgenai openai image "A flamingo in the pool" --image lounge.png --mask mask.png -o edited.png

# High fidelity (preserve facial features)
rawgenai openai image "Add sunglasses" --image portrait.png --fidelity high -o portrait_glasses.png

# Transparent background
rawgenai openai image "A red apple" -o apple.png --background transparent

# From file
rawgenai openai image --file prompt.txt -o output.png

# From stdin
echo "A forest at dawn" | rawgenai openai image -o forest.png

# JPEG with compression
rawgenai openai image "City skyline" -o city.jpeg --compression 80
```

## Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `--output` | `-o` | string | - | Yes | Output file path (format from extension) |
| `--image` | `-i` | string[] | - | No | Reference image(s), can be repeated (max 16) |
| `--mask` | - | string | - | No | Mask image for inpainting (PNG with alpha) |
| `--file` | - | string | - | No | Input prompt file |
| `--continue` | `-c` | string | - | No | Previous response ID for multi-turn conversation |
| `--model` | `-m` | string | `gpt-image-1` | No | Model name |
| `--size` | `-s` | string | `auto` | No | Image dimensions |
| `--quality` | `-q` | string | `auto` | No | Image quality |
| `--background` | - | string | `auto` | No | Background type |
| `--compression` | - | int | `100` | No | Compression 0-100 (JPEG/WebP only) |
| `--fidelity` | - | string | `low` | No | Input image fidelity (high/low) |
| `--moderation` | - | string | `auto` | No | Moderation level |

## Models

| Model | Description |
|-------|-------------|
| `gpt-image-1` | Standard GPT Image model (default) |
| `gpt-image-1.5` | State of the art, best quality |
| `gpt-image-1-mini` | Cost-effective, faster |

## Sizes

- `auto` (automatic, default)
- `1024x1024` (square)
- `1536x1024` (landscape)
- `1024x1536` (portrait)

## Quality

- `auto` (automatic, default)
- `high`
- `medium`
- `low`

## Output Formats

Format is determined by output file extension:

| Extension | Format | Notes |
|-----------|--------|-------|
| `.png` | PNG | Default, supports transparency |
| `.jpeg` / `.jpg` | JPEG | Supports compression |
| `.webp` | WebP | Supports compression and transparency |

## Background

| Value | Description |
|-------|-------------|
| `auto` | Automatic (default) |
| `opaque` | Solid background |
| `transparent` | Transparent (PNG/WebP only) |

## Input Fidelity

Controls how closely the output matches reference images (especially facial features).

| Value | Description |
|-------|-------------|
| `low` | Default, less strict matching |
| `high` | Preserve facial features and style closely |

## Mask Guidelines

- Mask must be PNG with alpha channel
- **Transparent areas** = regions to edit
- **Opaque areas** = regions to preserve
- Mask should match input image dimensions

## Output

```json
{
  "success": true,
  "file": "/path/to/output.png",
  "model": "gpt-image-1",
  "response_id": "resp_abc123"
}
```

Use `response_id` with `--continue` for multi-turn conversations.

## Errors

```json
{
  "success": false,
  "error": {
    "code": "missing_prompt",
    "message": "no prompt provided"
  }
}
```

### CLI Errors (before API call)

| Code | Description |
|------|-------------|
| `missing_api_key` | OPENAI_API_KEY not set |
| `missing_prompt` | No prompt provided |
| `missing_output` | --output flag not provided |
| `file_not_found` | Input file does not exist |
| `unsupported_format` | Output file extension not supported |
| `invalid_compression` | Compression value not in range 0-100 |
| `invalid_fidelity` | Fidelity value not 'high' or 'low' |
| `too_many_images` | More than 16 reference images |
| `transparent_requires_png_webp` | --background=transparent requires .png or .webp |
| `output_write_error` | Cannot write to output file |

### OpenAI API Errors

| HTTP | Code | Description |
|------|------|-------------|
| 400 | `invalid_prompt` | Prompt is invalid or too long |
| 400 | `content_policy` | Content violates policy |
| 400 | `invalid_request` | Other invalid request parameters |
| 401 | `invalid_api_key` | API key is invalid or revoked |
| 403 | `region_not_supported` | Region/country not supported |
| 429 | `rate_limit` | Too many requests |
| 429 | `quota_exceeded` | Quota/credits exhausted |
| 500 | `server_error` | OpenAI server error |
| 503 | `server_overloaded` | OpenAI server overloaded |

### Network Errors

| Code | Description |
|------|-------------|
| `connection_error` | Cannot connect to OpenAI API |
| `timeout` | Request timed out |
