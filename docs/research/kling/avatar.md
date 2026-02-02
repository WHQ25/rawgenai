# 数字人对口型

## 创建任务

| 网络协议 | 请求地址 | 请求方法 | 请求格式 | 响应格式 |
|---------|---------|---------|---------|---------|
| https | /v1/videos/avatar/image2video | POST | application/json | application/json |

## 请求头

| 字段 | 值 | 描述 |
|------|-----|------|
| Content-Type | application/json | 数据交换格式 |
| Authorization | 鉴权信息 | 鉴权信息 |

## 请求体

| 字段 | 类型 | 必填 | 默认值 | 描述 |
|------|------|------|--------|------|
| image | string | 是 | 无 | 数字人参考图，支持 Base64 或 URL |
| audio_id | string | 可选 | 空 | 通过试听接口生成的音频ID，与 sound_file 二选一 |
| sound_file | string | 可选 | 空 | 音频文件，支持 Base64 或 URL，与 audio_id 二选一 |
| prompt | string | 可选 | 空 | 正向文本提示词，可定义数字人动作、情绪及运镜等，不能超过2500个字符 |
| mode | string | 可选 | std | 生成模式。`std`：标准模式，`pro`：专家模式 |
| watermark_info | array | 可选 | 空 | 水印配置 |
| callback_url | string | 可选 | 无 | 回调通知地址 |
| external_task_id | string | 可选 | 空 | 自定义任务ID |

### 图片限制 (image)

- 支持 Base64 编码或 URL
- **Base64格式**：不要添加 `data:image/png;base64,` 前缀
- 格式支持：.jpg / .jpeg / .png
- 文件大小：≤10MB
- 宽高尺寸：≥300px
- 宽高比：1:2.5 ~ 2.5:1

### 音频限制

#### audio_id
- 通过试听接口生成的音频ID
- 仅支持30天内生成的音频
- 时长：2秒 ~ 300秒

#### sound_file
- 支持 Base64 编码或 URL
- 格式：.mp3 / .wav / .m4a / .aac
- 文件大小：≤5MB
- 时长：2秒 ~ 300秒

> **注**：audio_id 和 sound_file 二选一，不能同时为空，也不能同时有值

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
| https | /v1/videos/avatar/image2video/{id} | GET |

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
        "task_info": {
            "external_task_id": "string"
        },
        "task_result": {
            "videos": [
                {
                    "id": "string",
                    "url": "string",          // 30天后清理
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
| https | /v1/videos/avatar/image2video | GET |

### 查询参数

| 字段 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| pageNum | int | 1 | 页码，[1,1000] |
| pageSize | int | 30 | 每页数据量，[1,500] |
