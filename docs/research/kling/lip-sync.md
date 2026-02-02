# 对口型

## 人脸识别

| 网络协议 | 请求地址 | 请求方法 | 请求格式 | 响应格式 |
|---------|---------|---------|---------|---------|
| https | /v1/videos/identify-face | POST | application/json | application/json |

### 请求头

| 字段 | 值 | 描述 |
|------|-----|------|
| Content-Type | application/json | 数据交换格式 |
| Authorization | 鉴权信息 | 鉴权信息 |

### 请求体

| 字段 | 类型 | 必填 | 默认值 | 描述 |
|------|------|------|--------|------|
| video_id | string | 可选 | 无 | 通过可灵AI生成的视频ID，仅支持30天内生成的、时长≤60秒的视频。与 video_url 二选一 |
| video_url | string | 可选 | 无 | 视频URL。与 video_id 二选一 |

**视频限制**：
- 格式：.mp4 / .mov
- 文件大小：≤100MB
- 时长：2秒 ~ 60秒
- 分辨率：仅支持720p和1080p
- 边长：512px ~ 2160px

### 响应体

```json
{
    "code": 0,
    "message": "string",
    "request_id": "string",
    "data": {
        "session_id": "id",              // 会话ID，有效期24小时
        "final_unit_deduction": "string",
        "face_data": [
            {
                "face_id": "string",     // 视频中的人脸ID
                "face_image": "url",     // 人脸示意图
                "start_time": 0,         // 可对口型区间起点(ms)
                "end_time": 5200         // 可对口型区间终点(ms)
            }
        ]
    }
}
```

---

## 创建任务

| 网络协议 | 请求地址 | 请求方法 | 请求格式 | 响应格式 |
|---------|---------|---------|---------|---------|
| https | /v1/videos/advanced-lip-sync | POST | application/json | application/json |

### 请求头

| 字段 | 值 | 描述 |
|------|-----|------|
| Content-Type | application/json | 数据交换格式 |
| Authorization | 鉴权信息 | 鉴权信息 |

### 请求体

| 字段 | 类型 | 必填 | 默认值 | 描述 |
|------|------|------|--------|------|
| session_id | string | 是 | 无 | 会话ID，来自人脸识别接口返回 |
| face_choose | object[] | 是 | 无 | 人脸对口型配置 |
| watermark_info | array | 可选 | 空 | 水印配置 |
| external_task_id | string | 可选 | 无 | 自定义任务ID |
| callback_url | string | 可选 | 无 | 回调通知地址 |

#### face_choose 结构

| 字段 | 类型 | 必填 | 默认值 | 描述 |
|------|------|------|--------|------|
| face_id | string | 是 | 无 | 人脸ID（来自人脸识别接口） |
| audio_id | string | 可选 | 空 | 音频ID，与 sound_file 二选一 |
| sound_file | string | 可选 | 空 | 音频文件，与 audio_id 二选一 |
| sound_start_time | long | 是 | 无 | 音频裁剪起点(ms)，从0开始 |
| sound_end_time | long | 是 | 无 | 音频裁剪终点(ms) |
| sound_insert_time | long | 是 | 无 | 音频插入时间(ms)，从视频开始计算 |
| sound_volume | float | 可选 | 1 | 音频音量，[0, 2] |
| original_audio_volume | float | 可选 | 1 | 原始视频音量，[0, 2] |

**音频限制**：
- audio_id：30天内生成，2秒 ~ 60秒
- sound_file：格式.mp3/.wav/.m4a，≤5MB，2秒 ~ 60秒
- 音频和视频不能同时为空

**音频时间限制**：
- 裁剪后音频不得短于2秒
- 音频终点不得晚于原始音频总时长
- 插入音频与可对口型区间至少重合2秒
- 插入开始时间不得早于视频开始
- 插入结束时间不得晚于视频结束

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
| https | /v1/videos/advanced-lip-sync/{id} | GET |

### 请求路径参数

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| task_id | string | 是 | 对口型任务ID |

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
        "task_info": {
            "parent_video": {
                "id": "string",
                "url": "string",      // 30天后清理
                "duration": "string"
            },
            "external_task_id": "string"
        },
        "task_result": {
            "videos": [
                {
                    "id": "string",
                    "url": "string",
                    "watermark_url": "string",
                    "duration": "string"
                }
            ]
        },
        "watermark_info": {
            "enabled": boolean
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
| https | /v1/videos/advanced-lip-sync | GET |

### 查询参数

| 字段 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| pageNum | int | 1 | 页码，[1,1000] |
| pageSize | int | 30 | 每页数据量，[1,500] |
