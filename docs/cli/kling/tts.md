# Kling TTS Commands

Convert text to speech using Kling TTS.

---

## `kling tts`

Generate speech audio from text.

### Usage

```bash
rawgenai kling tts "Hello, world" --voice chat1_female_new-3 --language en -o hello.mp3
rawgenai kling tts "你好，世界" --voice chat1_female_new-3 -o hello.mp3
rawgenai kling tts "播放测试" --voice genshin_kirara --speak
```

### Flags

| Flag | Short | Type | Default | Required | Description |
|------|-------|------|---------|----------|-------------|
| `--output` | `-o` | string | | Yes* | Output file path |
| `--prompt-file` | `-f` | string | | No | Read text from file |
| `--voice` | | string | | Yes | Voice ID |
| `--language` | | string | `zh` | No | Voice language: `zh`, `en` |
| `--speed` | | float | 1.0 | No | Speech speed (0.8-2.0) |
| `--speak` | | bool | false | No | Play audio after generation |

*Output is required unless `--speak` is used.

### Available Voices (Bilingual)

These voices support both Chinese (`zh`) and English (`en`) via the `--language` flag:

| voice_id | 中文名称 | English Name | Category |
|----------|---------|--------------|----------|
| `genshin_vindi2` | 阳光少年 | Sunny | Young Male |
| `zhinen_xuesheng` | 懂事小弟 | Sage | Young Male |
| `ai_kaiya` | 阳光男生 | Shine | Young Male |
| `ai_chenjiahao_712` | 文艺小哥 | Lyric | Young Male |
| `ai_shatang` | 青春少女 | Blossom | Young Female |
| `genshin_klee2` | 温柔小妹 | Peppy | Young Female |
| `genshin_kirara` | 元气少女 | Dove | Young Female |
| `chat1_female_new-3` | 温柔姐姐 | Tender | Adult Female |
| `chengshu_jiejie` | 优雅贵妇 | Grace | Adult Female |
| `you_pingjing` | 温柔妈妈 | Helen | Adult Female |
| `ai_huangyaoshi_712` | 稳重老爸 | Rock | Adult Male |
| `ai_laoguowang_712` | 严肃上司 | Titan | Adult Male |
| `cartoon-boy-07` | 活泼男童 | Zippy | Child |
| `cartoon-girl-01` | 俏皮女童 | Sprite | Child |
| `laopopo_speech02` | 唠叨奶奶 | Prattle | Elderly |
| `heainainai_speech02` | 和蔼奶奶 | Hearth | Elderly |

### Output

```json
{
  "success": true,
  "task_id": "847487393472614443",
  "status": "succeed",
  "voice_id": "chat1_female_new-3",
  "duration": "3.276",
  "file": "/abs/path/hello.mp3"
}
```
