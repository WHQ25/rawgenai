# 视频生音效

## 创建任务

| 网络协议 | 请求地址 | 请求方法 | 请求格式 | 响应格式 |
|---------|---------|---------|---------|---------|
| https | /v1/audio/video-to-audio | POST | application/json | application/json |

## 请求头

| 字段 | 值 | 描述 |
|------|-----|------|
| Content-Type | application/json | 数据交换格式 |
| Authorization | 鉴权信息 | 鉴权信息 |

## 请求体

| 字段 | 类型 | 必填 | 默认值 | 描述 |
|------|------|------|--------|------|
| video_id | string | 可选 | 无 | 通过可灵AI生成的视频ID，与video_url二选一，仅支持30天内生成的3.0-20.0秒视频 |
| video_url | string | 可选 | 无 | 视频URL，与video_id二选一，格式MP4/MOV，文件≤100MB，时长3.0-20.0秒 |
| sound_effect_prompt | string | 可选 | 无 | 音效生成提示词，不超过200字符 |
| bgm_prompt | string | 可选 | 无 | 配乐生成提示词，不超过200字符 |
| asmr_mode | boolean | 可选 | false | 是否开启ASMR模式，true表示开启，增强细节音效，适合高沉浸内容 |
| external_task_id | string | 可选 | 无 | 自定义任务ID，单用户下需要保证唯一性 |
| callback_url | string | 可选 | 无 | 回调通知地址，服务端会在任务状态变更时主动通知 |

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
        "task_status": "string",  // submitted/processing/succeed/failed
        "created_at": 1722769557708,
        "updated_at": 1722769557708
    }
}
```

---

## 查询任务（单个）

| 网络协议 | 请求地址 | 请求方法 |
|---------|---------|---------|
| https | /v1/audio/video-to-audio/{id} | GET |

### 请求路径参数

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| task_id | string | 可选 | 视频生音频的任务ID |
| external_task_id | string | 可选 | 自定义任务ID，与task_id二选一 |

### 响应体

```json
{
    "code": 0,
    "message": "string",
    "request_id": "string",
    "data": {
        "task_id": "string",
        "task_status": "string",  // submitted/processing/succeed/failed
        "task_status_msg": "string",
        "final_unit_deduction": "string",
        "task_info": {
            "external_task_id": "string",
            "parent_video": {
                "id": "string",
                "url": "string",        // 30天后清理
                "duration": "string"
            }
        },
        "task_result": {
            "videos": [
                {
                    "id": "string",
                    "url": "string",
                    "duration": "string"
                }
            ],
            "audios": [
                {
                    "id": "string",
                    "url_mp3": "string",        // MP3格式音频URL（30天后清理）
                    "url_wav": "string",        // WAV格式音频URL（30天后清理）
                    "duration_mp3": "string",   // MP3总时长(s)
                    "duration_wav": "string"    // WAV总时长(s)
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
| https | /v1/audio/video-to-audio | GET |

### 查询参数

```
/v1/audio/video-to-audio?pageNum=1&pageSize=30
```

| 字段 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| pageNum | int | 1 | 页码，[1,1000] |
| pageSize | int | 30 | 每页数据量，[1,500] |

### 响应体

```json
{
    "code": 0,
    "message": "string",
    "request_id": "string",
    "data": [
        {
            "task_id": "string",
            "task_status": "string",
            "task_status_msg": "string",
            "final_unit_deduction": "string",
            "task_info": {
                "external_task_id": "string",
                "parent_video": {
                    "id": "string",
                    "url": "string",
                    "duration": "string"
                }
            },
            "task_result": {
                "videos": [
                    {
                        "id": "string",
                        "url": "string",
                        "duration": "string"
                    }
                ],
                "audios": [
                    {
                        "id": "string",
                        "url_mp3": "string",
                        "url_wav": "string",
                        "duration_mp3": "string",
                        "duration_wav": "string"
                    }
                ]
            },
            "created_at": 1722769557708,
            "updated_at": 1722769557708
        }
    ]
}
```
