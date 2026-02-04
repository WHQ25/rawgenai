# rawgenai dashscope image

Generate and edit images using Alibaba DashScope models. All calls are synchronous.

## Models

### Text-to-Image

| Model | Description |
|-------|-------------|
| `wan2.6-t2i` | Best image quality, flexible size (default) |
| `qwen-image-max` | Best realism and text rendering |
| `qwen-image-plus` | Multi-style, complex text rendering |

### Image Editing

| Model | Description | Input Images | Output Count |
|-------|-------------|-------------|-------------|
| `qwen-image-edit-max` | Best editing quality | 1-3 | 1-6 |
| `qwen-image-edit-plus` | Good editing, cost-effective | 1-3 | 1-6 |
| `qwen-image-edit` | Basic editing | 1-3 | 1 |
| `wan2.6-image` | Style transfer, subject consistency | 1-4 | 1-4 |

## Mode Auto-Selection

| Input | Default Model | Mode |
|-------|--------------|------|
| Prompt only | `wan2.6-t2i` | Text-to-image |
| `--image` + prompt | `qwen-image-edit-plus` | Image editing |

Users can override the default with `--model`.

---

## Usage

Generate or edit image(s). Returns image URLs directly (synchronous).

- **No `--image`** → text-to-image (default: wan2.6-t2i)
- **With `--image`** → image editing (default: qwen-image-edit-plus)

### Usage

```bash
# === Text-to-Image ===

# Default (wan2.6-t2i)
rawgenai dashscope image "一只猫在花园里"

# Auto-download
rawgenai dashscope image "一只猫在花园里" -o cat.png

# Multiple images
rawgenai dashscope image "一只猫在花园里" -n 4 -o cat.png

# Qwen-Image (best for text rendering)
rawgenai dashscope image "A poster with text: Hello World" -m qwen-image-max

# Custom size
rawgenai dashscope image "竖版海报" --size 960*1696

# === Image Editing ===

# Edit with default model (qwen-image-edit-plus)
rawgenai dashscope image "把猫变成狗" --image cat.jpg

# Edit with URL input
rawgenai dashscope image "Add sunglasses" --image https://example.com/face.jpg -o edited.png

# Multiple input images (e.g., style transfer)
rawgenai dashscope image "把图2的风格应用到图1" --image photo.jpg --image style.jpg

# Edit with wan2.6-image
rawgenai dashscope image "改为水彩风格" --image photo.jpg -m wan2.6-image

# Multiple output images
rawgenai dashscope image "生成多个变体" --image photo.jpg -n 4 -o variant.png

# From file
rawgenai dashscope image -f prompt.txt -o output.png

# From stdin
echo "一只猫" | rawgenai dashscope image -o cat.png
```

### Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `--model` | `-m` | string | auto | No | Model name (auto-selected by input) |
| `--image` | `-i` | string[] | | No | Input image(s) for editing (repeatable, local path or URL) |
| `--size` | `-s` | string | per model | No | Image size "width*height" |
| `--count` | `-n` | int | 1 | No | Number of output images |
| `--negative` | | string | | No | Negative prompt (max 500 chars) |
| `--seed` | | int | random | No | Random seed [0, 2147483647] |
| `--prompt-extend` | | bool | true | No | AI prompt enhancement |
| `--watermark` | | bool | false | No | Add AI-generated watermark |
| `--output` | `-o` | string | | No | Output file path (auto-download) |
| `--prompt-file` | `-f` | string | | No | Read prompt from file |

### Model-specific Constraints

| Constraint | Qwen-Image | Qwen-Image-Edit | Wan 2.6 T2I | Wan 2.6 Image |
|-----------|-----------|----------------|-----------|--------------|
| Max prompt | 800 chars | 800 chars | 2100 chars | 2000 chars |
| Input images | 0 | 1-3 | 0 | 1-4 |
| Output count (n) | 1 | 1-6 (max/plus), 1 (edit) | 1-4 | 1-4 |
| Negative prompt | Yes | Yes | Yes | Yes |
| Prompt extend | Yes | Yes (max/plus only) | Yes | Yes |

### Size Options

**Qwen-Image** (5 fixed sizes):

| Size | Ratio |
|------|-------|
| 1664*928 | 16:9 |
| 1472*1104 | 4:3 |
| 1328*1328 | 1:1 (default) |
| 1104*1472 | 3:4 |
| 928*1664 | 9:16 |

**Qwen-Image-Edit**: max/plus flexible [512, 2048] per dimension. Basic (edit): size ignored.

**Wan 2.6 T2I** (flexible, pixel area 1280²~1440², ratio 1:4~4:1):

| Size | Ratio |
|------|-------|
| 1696*960 | 16:9 |
| 1440*1088 | 4:3 |
| 1280*1280 | 1:1 (default) |
| 1088*1440 | 3:4 |
| 960*1696 | 9:16 |

**Wan 2.6 Image** (flexible, pixel area 768²~1280², ratio 1:4~4:1):

Default: auto based on input image. Common: `1280*1280`, `1024*1024`.

### Image Input Constraints

| Constraint | Qwen-Image-Edit | Wan 2.6 Image |
|-----------|----------------|--------------|
| Count | 1-3 | 1-4 |
| Formats | JPG, JPEG, PNG, BMP, TIFF, WEBP, GIF | JPG, JPEG, PNG, BMP, WEBP |
| Max resolution | 3072px per side | 5000px per side |
| Max file size | 10 MB | 10 MB |
| Input method | URL or base64 | URL or base64 |

### Output

Without `-o` (return URLs):
```json
{
  "success": true,
  "model": "wan2.6-t2i",
  "images": [
    {"url": "https://...", "index": 0}
  ]
}
```

With `-o` (auto-download, single image):
```json
{
  "success": true,
  "model": "wan2.6-t2i",
  "file": "/absolute/path/cat.png",
  "images": [
    {"url": "https://...", "index": 0}
  ]
}
```

With `-o` (auto-download, multiple images, suffix `_0`, `_1`...):
```json
{
  "success": true,
  "model": "wan2.6-t2i",
  "files": [
    "/absolute/path/cat_0.png",
    "/absolute/path/cat_1.png"
  ],
  "images": [
    {"url": "https://...", "index": 0},
    {"url": "https://...", "index": 1}
  ]
}
```

---

## API Details

### Endpoint

All models use the same sync endpoint:

```
POST {base_url}/services/aigc/multimodal-generation/generation
```

### Request Format

All models use messages format.

**Text-to-image:**
```json
{
  "model": "wan2.6-t2i",
  "input": {
    "messages": [{"role": "user", "content": [{"text": "prompt"}]}]
  },
  "parameters": {"size": "1280*1280", "n": 1, ...}
}
```

**Image editing:**
```json
{
  "model": "qwen-image-edit-plus",
  "input": {
    "messages": [{
      "role": "user",
      "content": [
        {"image": "https://example.com/photo.jpg"},
        {"image": "https://example.com/style.jpg"},
        {"text": "Apply style from image 2 to image 1"}
      ]
    }]
  },
  "parameters": {"n": 1, "size": "1280*1280", ...}
}
```

### Response Format

```json
{
  "output": {
    "choices": [{
      "finish_reason": "stop",
      "message": {
        "role": "assistant",
        "content": [
          {"image": "https://...png"},
          {"image": "https://...png"}
        ]
      }
    }]
  },
  "usage": {"image_count": 2, "size": "1280*1280"},
  "request_id": "xxx"
}
```

Image URLs valid for 24 hours.

---

## Validation Order

1. Prompt — required (positional arg, `--prompt-file`, or stdin)
2. `--image` + model compatibility:
   - If `--image` set with t2i model → error (`incompatible_image`)
   - If no `--image` with edit model → error (`missing_image`)
3. Image count — within model limit (qwen-edit: 1-3, wan2.6-image: 1-4)
4. Image file existence — skip for URLs
5. Model — must be in valid models list
6. Size — must be valid for the model
7. Count (n) — within model limit
8. Compatibility — prompt_extend not supported by qwen-image-edit (basic)
9. API key — `DASHSCOPE_API_KEY` required

## Error Codes

| Code | Description |
|------|-------------|
| `missing_prompt` | No prompt provided |
| `missing_image` | Edit model requires `--image` |
| `missing_api_key` | DASHSCOPE_API_KEY not set |
| `invalid_model` | Unknown model name |
| `invalid_size` | Invalid size for model |
| `invalid_count` | Count out of range for model |
| `image_not_found` | Local image file not found |
| `too_many_images` | Too many input images for model |
| `incompatible_image` | `--image` not supported by t2i model |
| `incompatible_prompt_extend` | prompt_extend not supported by model |
| `request_error` | HTTP request failed |
| `api_error` | DashScope API error |
| `download_error` | File download failed |
| `write_error` | Cannot write file |
