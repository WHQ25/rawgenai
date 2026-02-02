# 视频延长

## 概述

- 视频延长是指对文生/图生视频结果进行时间上的延长
- 单次可延长4~5秒
- 使用的模型和模式与源视频相同
- 被延长后的视频可以再次延长，但总时长不能超过3分钟
- 原视频需要在30天内生成（30天后会被清理无法延长）

---

## 创建任务

| 网络协议 | 请求地址 | 请求方法 | 请求格式 | 响应格式 |
|---------|---------|---------|---------|---------|
| https | /v1/videos/video-extend | POST | application/json | application/json |

## 请求头

| 字段 | 值 | 描述 |
|------|-----|------|
| Content-Type | application/json | 数据交换格式 |
| Authorization | 鉴权信息 | 鉴权信息 |

## 请求体

| 字段 | 类型 | 必填 | 默认值 | 描述 |
|------|------|------|--------|------|
| video_id | string | 是 | 无 | 视频ID（支持文生/图生/视频延长生成的视频） |
| prompt | string | 可选 | 无 | 正向文本提示词，不能超过2500个字符 |
| negative_prompt | string | 可选 | 无 | 负向文本提示词，不能超过2500个字符 |
| cfg_scale | float | 可选 | 0.5 | 提示词参考强度。取值范围：[0, 1]，数值越大参考强度越大 |
| watermark_info | array | 可选 | 空 | 水印配置 |
| callback_url | string | 可选 | 无 | 回调通知地址 |
| external_task_id | string | 可选 | 无 | 自定义任务ID |

## 响应体

```json
{
    "code": 0,
    "message": "string",
    "request_id": "string",
    "data": {
        "task_id": "string",
        "task_status": "string",  // submitted/processing/succeed/failed
        "task_info": {
            "external_task_id": "string"
        },
        "created_at": 1722769557708,
        "updated_at": 1722769557708
    }
}
```

---

## 查询任务（单个）

| 网络协议 | 请求地址 | 请求方法 |
|---------|---------|---------|
| https | /v1/videos/video-extend/{id} | GET |

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
            "parent_video": {
                "id": "string",         // 延长前的视频ID
                "url": "string",        // 延长前视频的URL（30天后清理）
                "duration": "string"    // 延长前的视频总时长(s)
            },
            "external_task_id": "string"
        },
        "task_result": {
            "videos": [
                {
                    "id": "string",           // 延长后的完整视频ID
                    "url": "string",          // 延长后视频的URL
                    "watermark_url": "string",// 含水印视频URL
                    "duration": "string"      // 视频总时长(s)
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
| https | /v1/videos/video-extend | GET |

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
            "task_info": {
                "parent_video": {
                    "id": "string",
                    "url": "string",
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
    ]
}
```
