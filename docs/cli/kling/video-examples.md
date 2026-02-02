# Kling Video Examples

Common usage patterns and workflows.

---

## Text-to-video

```bash
# Simple generation
kling video create "A golden retriever running on the beach"

# With options
kling video create "Cinematic shot of a mountain landscape" \
  --mode pro \
  --duration 10 \
  --ratio 16:9

# Check status and download
kling video status <task_id>
kling video download <task_id> -o beach_dog.mp4
```

---

## First Frame / Last Frame

```bash
# First frame only
kling video create "The character waves hello" \
  --first-frame portrait.png \
  --duration 5

# First + last frame (morphing transition)
kling video create "Smooth morphing transition" \
  --first-frame scene1.png \
  --last-frame scene2.png \
  --duration 7
```

---

## Reference Images

```bash
# Single reference (character)
kling video create "<<<image_1>>> walks through a forest" \
  --ref-image character.png \
  --ratio 16:9 \
  --duration 5

# Multiple references (character + style)
kling video create "<<<image_1>>> in the style of <<<image_2>>>" \
  --ref-image character.png \
  --ref-image art_style.png \
  --ratio 1:1 \
  --duration 7

# First frame + reference images
kling video create "<<<image_1>>> appears and joins the scene" \
  --first-frame scene.png \
  --ref-image guest_character.png \
  --duration 5
```

---

## Custom Elements

```bash
# Create element first
kling video element create "MyCat" \
  -d "Orange tabby cat with green eyes" \
  -f cat_front.jpg \
  -r cat_side.jpg -r cat_back.jpg

# Use custom element
kling video create "<<<element_1>>> walking in a garden, happy" \
  --element 123456789 \
  --ratio 16:9 \
  --duration 5

# Multiple elements
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

---

## Reference Video (Style/Camera)

```bash
# Copy camera movement (keeps original sound by default)
kling video create "Same camera movement, but with a cat" \
  --ref-video https://example.com/cool_camera.mp4 \
  --ratio 16:9 \
  --duration 5

# Generate next shot
kling video create "Continue from <<<video_1>>>, generate next shot" \
  --ref-video https://example.com/previous.mp4

# Without original audio
kling video create "Same style as <<<video_1>>>" \
  --ref-video https://example.com/reference.mp4 \
  --ref-exclude-sound
```

---

## Video Editing (Base Video)

```bash
# Add element (keeps original sound by default)
kling video create "Add sunglasses to the person in <<<video_1>>>" \
  --base-video https://example.com/person.mp4

# Replace element (without original sound)
kling video create "Replace the car in <<<video_1>>> with a bicycle" \
  --base-video https://example.com/street.mp4 \
  --ref-exclude-sound

# Change style
kling video create "Make <<<video_1>>> look like anime style" \
  --base-video https://example.com/realistic.mp4

# Add element from reference image
kling video create "Put <<<image_1>>> hat on the person in <<<video_1>>>" \
  --base-video https://example.com/person.mp4 \
  --ref-image hat.png
```

---

## Video Extension

```bash
# Create initial video
kling video create "A spaceship takes off" --duration 5
# Get video_id from status

# Extend the video (legacy models only)
kling video extend <video_id> --prompt "The spaceship enters hyperspace"

# Download extended video
kling video status <extend_task_id> --type extend
kling video download <extend_task_id> --type extend -o spaceship_full.mp4
```

---

## Add Sound Effects

```bash
# Add sound to completed video
kling video add-sound <video_id> \
  --sound "rocket engine roaring" \
  --bgm "epic orchestral music"

# Check status
kling video status <sound_task_id> --type add-sound -v

# Download video with sound
kling video download <sound_task_id> --type add-sound -o spaceship_with_sound.mp4

# Download audio only
kling video download <sound_task_id> --type add-sound --format mp3 -o audio.mp3
```

---

## Legacy Models

### Text-to-Video

```bash
# Using legacy model with camera control
kling video create-from-text "A bird flying over mountains" \
  --model kling-v2-6 \
  --camera-control '{"type":"simple","config":{"zoom":-3}}' \
  --sound
```

### Image-to-Video with Motion Control

```bash
# Static mask (keep background still)
kling video create-from-image "The person walks forward" \
  -i person.png \
  --static-mask background_mask.png \
  --model kling-v2-master

# Dynamic mask (control movement path)
kling video create-from-image "The ball bounces" \
  -i ball_scene.png \
  --dynamic-mask '[{"mask":"ball_mask.png","trajectories":[[{"x":100,"y":300},{"x":150,"y":100},{"x":200,"y":300}]]}]' \
  --model kling-v2-master
```

---

## Complete Workflow Example

```bash
# 1. Create a custom element
kling video element create "Hero" \
  -d "A superhero in red cape" \
  -f hero_front.jpg \
  -r hero_side.jpg

# 2. Generate video with element
kling video create "<<<element_1>>> flying through clouds" \
  --element 123456789 \
  --mode pro \
  --duration 5 \
  --ratio 16:9

# 3. Wait for completion
kling video status <task_id>
# Repeat until status is "succeed"

# 4. Add sound effects
kling video add-sound <video_id> \
  --sound "wind whooshing, cape flapping" \
  --bgm "heroic orchestral theme"

# 5. Wait for sound processing
kling video status <sound_task_id> --type add-sound

# 6. Download final video
kling video download <sound_task_id> --type add-sound -o hero_flying.mp4
```
