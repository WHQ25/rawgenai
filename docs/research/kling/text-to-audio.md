# 文生音效

## 创建任务

| 网络协议 | 请求地址 | 请求方法 | 请求格式 | 响应格式 |
|---------|---------|---------|---------|---------|
| https | /v1/audio/text-to-audio | POST | application/json | application/json |

## 请求头

| 字段 | 值 | 描述 |
|------|-----|------|
| Content-Type | application/json | 数据交换格式 |
| Authorization | 鉴权信息 | 鉴权信息 |

## 请求体

| 字段 | 类型 | 必填 | 默认值 | 描述 |
|------|------|------|--------|------|
| prompt | string | 是 | 无 | 文本提示词，内容长度不超过200字符 |
| duration | float | 是 | 无 | 生成音频的时长，取值范围：3.0秒至10.0秒，支持小数点后一位精度 |
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
| https | /v1/audio/text-to-audio/{id} | GET |

### 请求路径参数

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| task_id | string | 可选 | 文生音频的任务ID |
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
        "task_info": {
            "external_task_id": "string"
        },
        "task_result": {
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
        "final_unit_deduction": "string",
        "created_at": 1722769557708,
        "updated_at": 1722769557708
    }
}
```

---

## 查询任务（列表）

| 网络协议 | 请求地址 | 请求方法 |
|---------|---------|---------|
| https | /v1/audio/text-to-audio | GET |

### 查询参数

```
/v1/audio/text-to-audio?pageNum=1&pageSize=30
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
            "task_info": {
                "external_task_id": "string"
            },
            "task_result": {
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
            "final_unit_deduction": "string",
            "created_at": 1722769557708,
            "updated_at": 1722769557708
        }
    ]
}
```
