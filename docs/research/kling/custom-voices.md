# 自定义音色

## 创建自定义音色

| 网络协议 | 请求地址 | 请求方法 | 请求格式 | 响应格式 |
|---------|---------|---------|---------|---------|
| https | /v1/general/custom-voices | POST | application/json | application/json |

## 请求头

| 字段 | 值 | 描述 |
|------|-----|------|
| Content-Type | application/json | 数据交换格式 |
| Authorization | 鉴权信息 | 鉴权信息 |

## 请求体

| 字段 | 类型 | 必填 | 默认值 | 描述 |
|------|------|------|--------|------|
| voice_name | string | 是 | 无 | 音色名称，不超过20字符 |
| voice_url | string | 可选 | 空 | 音色数据文件URL，支持.mp3/.wav/.mp4/.mov，5-30秒，干净无杂音，仅一种人声 |
| video_id | string | 可选 | 空 | 历史作品ID，支持V2.6版本模型生成的视频、数字人API视频、对口型API视频 |
| callback_url | string | 可选 | 空 | 回调通知地址 |
| external_task_id | string | 可选 | 空 | 自定义任务ID，单用户下需要保证唯一性 |

**注意**：
- voice_url和video_id二选一
- 音频需要干净无杂音，有且只能有一种人声
- 时长不短于5秒且不长于30秒
- 仅满足以下条件的视频可用于定制音色：
  - 使用V2.6版本模型生成且开启sound参数为on的视频
  - 通过数字人API生成的视频
  - 通过对口型API生成的视频

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

## 查询自定义音色（单个）

| 网络协议 | 请求地址 | 请求方法 |
|---------|---------|---------|
| https | /v1/general/custom-voices/{id} | GET |

### 请求路径参数

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| task_id | string | 必填 | 生成音色的任务ID |
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
            "external_task_id": "string"
        },
        "task_result": {
            "voices": [
                {
                    "voice_id": "string",
                    "voice_name": "string",
                    "trial_url": "string",
                    "owned_by": "string"  // kling为官方音色库，其他为创作者ID
                }
            ]
        },
        "created_at": 1722769557708,
        "updated_at": 1722769557708
    }
}
```

---

## 查询自定义音色（列表）

| 网络协议 | 请求地址 | 请求方法 |
|---------|---------|---------|
| https | /v1/general/custom-voices | GET |

### 查询参数

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
                "external_task_id": "string"
            },
            "task_result": {
                "voices": [
                    {
                        "voice_id": "string",
                        "voice_name": "string",
                        "trial_url": "string",
                        "owned_by": "string"
                    }
                ]
            },
            "created_at": 1722769557708,
            "updated_at": 1722769557708
        }
    ]
}
```

---

## 查询官方音色（列表）

| 网络协议 | 请求地址 | 请求方法 |
|---------|---------|---------|
| https | /v1/general/presets-voices | GET |

### 查询参数

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
            "task_result": {
                "voices": [
                    {
                        "voice_id": "string",
                        "voice_name": "string",
                        "trial_url": "string",
                        "owned_by": "kling"
                    }
                ]
            },
            "created_at": 1722769557708,
            "updated_at": 1722769557708
        }
    ]
}
```

---

## 删除自定义音色

| 网络协议 | 请求地址 | 请求方法 | 请求格式 | 响应格式 |
|---------|---------|---------|---------|---------|
| https | /v1/general/delete-voices | POST | application/json | application/json |

### 请求头

| 字段 | 值 | 描述 |
|------|-----|------|
| Content-Type | application/json | 数据交换格式 |
| Authorization | 鉴权信息 | 鉴权信息 |

### 请求体

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| voice_id | string | 是 | 待删除的音色ID，仅支持删除自定义音色 |

### 响应体

```json
{
    "code": 0,
    "message": "string",
    "request_id": "string",
    "data": {
        "task_id": "string",
        "task_status": "string"  // submitted/processing/succeed/failed
    }
}
```
