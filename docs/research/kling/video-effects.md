# 视频特效

## 创建任务

| 网络协议 | 请求地址 | 请求方法 | 请求格式 | 响应格式 |
|---------|---------|---------|---------|---------|
| https | /v1/videos/effects | POST | application/json | application/json |

## 请求头

| 字段 | 值 | 描述 |
|------|-----|------|
| Content-Type | application/json | 数据交换格式 |
| Authorization | 鉴权信息 | 鉴权信息 |

## 请求体

| 字段 | 类型 | 必填 | 默认值 | 描述 |
|------|------|------|--------|------|
| effect_scene | string | 是 | 无 | 场景名称（共198款特效） |
| input | object | 是 | 无 | 根据特效类型的输入结构 |
| watermark_info | array | 可选 | 空 | 水印配置 |
| callback_url | string | 可选 | 无 | 回调通知地址 |
| external_task_id | string | 可选 | 无 | 自定义任务ID |

## 特效类型

### 单图特效（193款）

**示例场景**：pet_lion, pet_vlogger, crystal_horse, drunk_dance, bouncy_dance, new_year_greeting, prosperity, fortune_god_transform等

**input 结构**：
```json
{
    "effect_scene": "pet_lion",
    "input": {
        "image": "https://example.com/image.jpg",
        "duration": "5"
    }
}
```

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| image | string | 是 | 参考图像，支持 Base64 或 URL |
| duration | string | 是 | 视频时长（秒）。枚举值：`5`, `10` |

### 双人互动特效（5款）

**场景**：cheers_2026, kiss_pro, fight_pro, hug_pro, heart_gesture_pro

**特性**：
- 包含合照功能
- 自适应拼接两张人物图为合照
- 第一张图在左边，第二张图在右边

**input 结构**：
```json
{
    "effect_scene": "hug_pro",
    "input": {
        "images": [
            "https://example.com/image1.jpg",
            "https://example.com/image2.jpg"
        ],
        "duration": "5"
    }
}
```

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| images | string[] | 是 | 参考图像数组，长度必须为2 |
| duration | string | 是 | 视频时长。仅支持 `5` |

### 图片限制

- 支持 Base64 编码或 URL
- **Base64格式**：不要添加 `data:image/png;base64,` 前缀
- 格式：.jpg / .jpeg / .png
- 文件大小：≤10MB
- 宽高尺寸：≥300px
- 宽高比：1:2.5 ~ 2.5:1

## 模型和模式支持

| 特效类型 | 模型 | 模式 |
|---------|------|------|
| cheers_2026, kiss_pro, fight_pro, hug_pro | 不需要填 | 不需要填 |
| fight | kling-v1-6 | 需要填 |
| hug, kiss, heart_gesture | kling-v1, kling-v1-5, kling-v1-6 | std / pro |

## 响应体

```json
{
    "code": 0,
    "message": "string",
    "request_id": "string",
    "data": {
        "task_id": "string",
        "task_info": {
            "external_task_id": "string"
        },
        "task_status": "string",
        "created_at": 1722769557708,
        "updated_at": 1722769557708
    }
}
```

---

## 查询任务（单个）

| 网络协议 | 请求地址 | 请求方法 |
|---------|---------|---------|
| https | /v1/videos/effects/{id} | GET |

### 请求路径参数

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| task_id | string | 是 | 任务ID（与 external_task_id 二选一） |
| external_task_id | string | 可选 | 自定义任务ID |

### 响应体

```json
{
    "code": 0,
    "message": "string",
    "request_id": "string",
    "data": {
        "task_id": "string",
        "task_status": "string",
        "task_status_msg": "string",
        "final_unit_deduction": "string",
        "watermark_info": {
            "enabled": boolean
        },
        "task_info": {
            "external_task_id": "string"
        },
        "task_result": {
            "videos": [
                {
                    "id": "string",
                    "url": "string",      // 30天后清理
                    "duration": "string"
                }
            ]
        },
        "created_at": 1722769557708,
        "updated_at": 1722769557708
    }
}
```

---

## 查询任务（列表）

| 网络协议 | 请求地址 | 请求方法 |
|---------|---------|---------|
| https | /v1/videos/effects | GET |

### 查询参数

```
/v1/videos/effects?pageNum=1&pageSize=30
```

| 字段 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| pageNum | int | 1 | 页码，[1,1000] |
| pageSize | int | 30 | 每页数据量，[1,500] |

## 特效场景列表

**单图特效（193款）**：fortune_god_transform, new_year_feast, ring_in_new, horse_year_firework, pet_vlogger, crystal_horse, lateral_shift_transition, drunk_dance, drunk_dance_pet, daoma_dance, bouncy_dance, smooth_sailing_dance, new_year_greeting, lion_dance, prosperity, great_success, golden_horse_fortune, red_packet_box, lucky_horse_year, lucky_red_packet, lucky_money_come, lion_dance_pet, dumpling_making_pet, fish_making_pet, pet_red_packet, lantern_glow, expression_challenge, overdrive, heart_gesture_dance, poping, martial_arts, running, nezha, motorcycle_dance, subject_3_dance, ghost_step_dance, phantom_jewel, zoom_out, dollar_rain_pro, pet_bee_pro, countdown_teleport, santa_random_surprise, magic_match_tree, bullet_time_360, happy_birthday, birthday_star, thumbs_up_pro, tiger_hug_pro, pet_lion_pro, surprise_bouquet, bouquet_drop, 3d_cartoon_1_pro, firework_2026, glamour_photo_shoot, box_of_joy, first_toast_of_the_year, my_santa_pic, santa_gift, steampunk_christmas, snowglobe, christmas_photo_shoot, ornament_crash, santa_express, instant_christmas, particle_santa_surround, coronation_of_frost, building_sweater, spark_in_the_snow, scarlet_and_snow, cozy_toon_wrap, bullet_time_lite, magic_cloak, balloon_parade, jumping_ginger_joy, bullet_time, c4d_cartoon_pro, pure_white_wings, black_wings, golden_wing, pink_pink_wings, venomous_spider, throne_of_king, luminous_elf, woodland_elf, japanese_anime_1, american_comics, guardian_spirit, swish_swish, snowboarding, witch_transform, vampire_transform, pumpkin_head_transform, demon_transform, mummy_transform, zombie_transform, cute_pumpkin_transform, cute_ghost_transform, knock_knock_halloween, halloween_escape, baseball, inner_voice, a_list_look, memory_alive, trampoline, trampoline_night, pucker_up, guess_what, feed_mooncake, rampage_ape, flyer, dishwasher, pet_chinese_opera, magic_fireball, gallery_ring, pet_moto_rider, muscle_pet, squeeze_scream, pet_delivery, running_man, disappear, mythic_style, steampunk, 3d_cartoon_2, eagle_snatch, hug_from_past, firework, media_interview, pet_chef, santa_gifts, santa_hug, girlfriend, boyfriend, heart_gesture_1, pet_wizard, smoke_smoke, instant_kid, dollar_rain, cry_cry, building_collapse, gun_shot, mushroom, double_gun, pet_warrior, lightning_power, jesus_hug, shark_alert, long_hair, lie_flat, polar_bear_hug, brown_bear_hug, jazz_jazz, office_escape_plow, fly_fly, watermelon_bomb, pet_dance, boss_coming, wool_curly, pet_bee, marry_me, swing_swing, day_to_night, piggy_morph, wig_out, car_explosion, ski_ski, siblings, construction_worker, let's_ride, snatched, magic_broom, felt_felt, jumpdrop, splashsplash, surfsurf, fairy_wing, angel_wing, dark_wing, skateskate, plushcut, jelly_press, jelly_slice, jelly_squish, jelly_jiggle, pixelpixel, yearbook, instant_film, anime_figure, rocketrocket, bloombloom, dizzydizzy, fuzzyfuzzy, squish, expansion

**双人互动特效（5款）**：cheers_2026, kiss_pro, fight_pro, hug_pro, heart_gesture_pro

更多参数详见：特效模版中心
