# Kling Voice Commands

Commands for managing custom voices for use in video generation.

---

## `kling voice create`

Create a custom voice from audio or video.

### Usage

```bash
# From audio file (local)
kling voice create "MyVoice" --audio voice_sample.mp3

# From audio URL
kling voice create "MyVoice" --audio "https://example.com/audio.mp3"

# From video ID (v2.6/avatar/lip-sync generated video)
kling voice create "MyVoice" --video-id video_123
```

### Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `--audio` | `-a` | string | | No* | Audio file for voice cloning (local file or URL) |
| `--video-id` | | string | | No* | Video ID from v2.6/avatar/lip-sync generation |

*Either `--audio` or `--video-id` is required, but not both.

### Audio Requirements

- Duration: 5-30 seconds
- Clean audio without background noise
- Single speaker only
- Formats: MP3, WAV, MP4, MOV

### Video Requirements

Only the following videos can be used for voice cloning:
- Videos generated with v2.6 model and `--sound` enabled
- Videos generated via avatar API (`create-avatar`)
- Videos generated via lip-sync API

### Output

```json
{
  "success": true,
  "task_id": "xxx",
  "status": "submitted",
  "voice_name": "MyVoice"
}
```

---

## `kling voice status`

Get voice creation status.

### Usage

```bash
kling voice status <task_id>
```

### Output (Processing)

```json
{
  "success": true,
  "task_id": "xxx",
  "status": "processing"
}
```

### Output (Completed)

```json
{
  "success": true,
  "task_id": "xxx",
  "status": "succeed",
  "voice_id": "voice_xxx",
  "voice_name": "MyVoice",
  "trial_url": "https://..."
}
```

---

## `kling voice list`

List custom or official voices.

### Usage

```bash
# List custom voices (default)
kling voice list

# List official voices
kling voice list --type official

# With pagination
kling voice list --type custom --limit 50 --page 2
```

### Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `--type` | `-t` | string | `custom` | No | Voice type: custom, official |
| `--limit` | `-l` | int | `30` | No | Maximum voices to return (1-500) |
| `--page` | `-p` | int | `1` | No | Page number |

### Output

```json
{
  "success": true,
  "type": "custom",
  "voices": [
    {
      "voice_id": "voice_xxx",
      "voice_name": "MyVoice",
      "trial_url": "https://...",
      "owned_by": "user_id"
    }
  ],
  "count": 1
}
```

**Note:** For official voices, `owned_by` will be `"kling"`.

---

## `kling voice delete`

Delete a custom voice.

### Usage

```bash
kling voice delete <voice_id>
```

### Output

```json
{
  "success": true,
  "voice_id": "voice_xxx",
  "status": "deleted"
}
```

---

## Using Voices in Video Generation

After creating a voice, use it in video generation with the `--voice` flag:

```bash
# Use in image-to-video (v2.6 only)
kling video create-from-image "The person speaks" \
  -i photo.png \
  --model kling-v2-6 \
  --voice voice_xxx

# Use multiple voices
kling video create-from-image "Two people talking" \
  -i photo.png \
  --model kling-v2-6 \
  --voice voice_xxx,voice_yyy
```

---

## Error Codes

### CLI Errors

| Code | Description |
|------|-------------|
| `missing_name` | Voice name is required |
| `invalid_name` | Voice name exceeds 20 characters |
| `missing_audio` | Audio file or video ID is required |
| `conflicting_source` | Cannot use both --audio and --video-id |
| `audio_not_found` | Audio file not found |
| `audio_read_error` | Cannot read audio file |
| `missing_task_id` | Task ID is required |
| `missing_voice_id` | Voice ID is required |
| `invalid_type` | Invalid voice type (use custom or official) |
| `invalid_limit` | Limit must be between 1 and 500 |
| `invalid_page` | Page must be at least 1 |
| `voice_failed` | Voice creation failed |
