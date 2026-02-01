# grok image

Generate or edit images using xAI Grok API.

## Usage

```bash
# Generate an image (no --image flag)
rawgenai grok image "A futuristic city at sunset" -o city.png

# Edit an image (with --image flag)
rawgenai grok image "Make the sky purple" -i city.png -o city_purple.png

# Generate with options
rawgenai grok image "A cute robot" -o robot.png -n 4 -a 16:9

# From file
rawgenai grok image --prompt-file prompt.txt -o output.png

# From stdin
cat prompt.txt | rawgenai grok image -o output.png
```

## Modes

| Mode | Trigger | API Endpoint |
|------|---------|--------------|
| Generate | No `--image` flag | `POST /v1/images/generations` |
| Edit | With `--image` flag | `POST /v1/images/edits` |

## Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `--output` | `-o` | string | | Yes | Output file path (.png, .jpeg, .jpg) |
| `--prompt-file` | | string | | No | Read prompt from file |
| `--image` | `-i` | string | | No | Input image for edit mode |
| `--n` | `-n` | int | 1 | No | Number of images (1-10, generate only) |
| `--aspect` | `-a` | string | "1:1" | No | Aspect ratio (generate only) |

### Aspect Ratios (Generate Mode)

- `1:1` (default)
- `16:9`
- `9:16`
- `4:3`
- `3:4`

### Supported Formats

- `.png`
- `.jpeg`
- `.jpg`

## Output Format

### Success

```json
{
  "success": true,
  "file": "/absolute/path/to/output.png",
  "mode": "generate"
}
```

```json
{
  "success": true,
  "file": "/absolute/path/to/output.png",
  "mode": "edit"
}
```

### Error

```json
{
  "success": false,
  "error": {
    "code": "error_code",
    "message": "Error description"
  }
}
```

## Error Codes

### CLI Errors

| Code | Description |
|------|-------------|
| `missing_prompt` | No prompt provided |
| `missing_output` | No output file specified |
| `unsupported_format` | Output format not supported |
| `invalid_n` | n not in range 1-10 |
| `invalid_aspect` | Invalid aspect ratio |
| `image_not_found` | Input image file not found |
| `missing_api_key` | XAI_API_KEY not set |

### API Errors

| Code | Description |
|------|-------------|
| `invalid_request` | Bad request (400) |
| `invalid_api_key` | Invalid or revoked API key (401) |
| `permission_denied` | Insufficient permissions (403) |
| `rate_limit` | Too many requests (429) |
| `quota_exceeded` | API quota exhausted |
| `server_error` | xAI server error (500) |
| `server_overloaded` | Server overloaded (503) |
| `timeout` | Request timed out |
| `connection_error` | Cannot connect to API |

## Environment Variables

| Variable | Description |
|----------|-------------|
| `XAI_API_KEY` | xAI API key (required) |
