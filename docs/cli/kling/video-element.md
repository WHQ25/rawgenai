# Kling Video Element Commands

Commands for managing custom elements (characters, objects, styles) for use in video generation.

---

## `kling video element create`

Create a custom element for use in video generation.

### Usage

```bash
# Create element with required fields
kling video element create "MyCharacter" \
  -d "A cute cartoon cat with orange fur" \
  -f frontal.jpg \
  -r side.jpg -r back.jpg

# With tags
kling video element create "MyPet" \
  -d "My pet dog" \
  -f dog_front.jpg \
  -r dog_side.jpg \
  -t o_103
```

### Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `--description` | `-d` | string | | Yes | Element description (max 100 chars) |
| `--frontal` | `-f` | string | | Yes | Frontal reference image path |
| `--ref` | `-r` | string[] | | Yes | Additional reference images (1-3) |
| `--tag` | `-t` | string[] | | No | Element tags |

### Tags

| Tag ID | Name | Description |
|--------|------|-------------|
| `o_101` | 热梗 | Internet memes |
| `o_102` | 人物 | Characters |
| `o_103` | 动物 | Animals |
| `o_104` | 道具 | Props |
| `o_105` | 服饰 | Clothing |
| `o_106` | 场景 | Scenes |
| `o_107` | 特效 | Effects |
| `o_108` | 其他 | Other |

### Image Requirements

- Formats: JPEG, PNG
- Size: ≤10MB per image
- Dimensions: ≥300px
- Aspect ratio: 1:2.5 ~ 2.5:1

### Output

```json
{
  "success": true,
  "element_id": 123456789,
  "element_name": "MyCharacter"
}
```

---

## `kling video element list`

List custom or official elements.

### Usage

```bash
# List custom elements (default)
kling video element list

# List official elements
kling video element list --type official

# With pagination
kling video element list --type custom --limit 50 --page 2
```

### Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `--type` | `-t` | string | `custom` | No | Element type: custom, official |
| `--limit` | `-l` | int | `30` | No | Maximum elements to return (1-500) |
| `--page` | `-p` | int | `1` | No | Page number |

### Output

```json
{
  "success": true,
  "type": "custom",
  "elements": [
    {
      "element_id": 123456789,
      "name": "MyCharacter",
      "description": "A cute cartoon cat",
      "frontal_url": "https://...",
      "owned_by": "user_id"
    }
  ],
  "count": 1
}
```

---

## `kling video element delete`

Delete a custom element.

### Usage

```bash
kling video element delete <element_id>
```

### Output

```json
{
  "success": true,
  "element_id": "123456789",
  "status": "deleted"
}
```

---

## Using Elements in Video Generation

After creating an element, use it in video generation with the `--element` flag:

```bash
# Use single element
kling video create "<<<element_1>>> walking in a garden, happy" \
  --element 123456789 \
  --ratio 16:9 \
  --duration 5

# Use multiple elements
kling video create "<<<element_1>>> meets <<<element_2>>> in the park" \
  --element 123456789 \
  --element 987654321 \
  --ratio 16:9 \
  --duration 7

# Combine element with reference image
kling video create "<<<element_1>>> wearing the outfit from <<<image_1>>>" \
  --element 123456789 \
  --ref-image outfit.png \
  --ratio 16:9 \
  --duration 5
```

**Note:** Elements are referenced using `<<<element_N>>>` placeholders in the prompt, where N corresponds to the order of `--element` flags.
