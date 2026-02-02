# 图像生成

## 创建任务

| 网络协议 | 请求地址 | 请求方法 | 请求格式 | 响应格式 |
|---------|---------|---------|---------|---------|
| https | /v1/images/generations | POST | application/json | application/json |

## 请求头

| 字段 | 值 | 描述 |
|------|-----|------|
| Content-Type | application/json | 数据交换格式 |
| Authorization | 鉴权信息 | 鉴权信息 |

## 请求体

| 字段 | 类型 | 必填 | 默认值 | 描述 |
|------|------|------|--------|------|
| model_name | string | 可选 | kling-v1 | 模型名称，枚举值：kling-v1, kling-v1-5, kling-v2, kling-v2-new, kling-v2-1 |
| prompt | string | 是 | 无 | 正向文本提示词，不超过2500字符 |
| negative_prompt | string | 可选 | 空 | 负向文本提示词，不超过2500字符（图生图场景下不支持） |
| image | string | 可选 | 空 | 参考图像（Base64编码或URL），支持.jpg/.jpeg/.png，≤10MB，≥300px，宽高比1:2.5~2.5:1 |
| image_reference | string | 可选 | 无 | 图片参考类型，枚举值：subject（角色特征参考）、face（人物长相参考），仅kling-v1-5支持 |
| image_fidelity | float | 可选 | 0.5 | 生成过程中对上传图片的参考强度，[0,1]，数值越大参考强度越大 |
| human_fidelity | float | 可选 | 0.45 | 面部参考强度（人物五官相似度），仅image_reference为subject时生效，[0,1] |
| resolution | string | 可选 | 1k | 生成图片清晰度，枚举值：1k（1K标清）、2k（2K高清），不同模型版本支持范围不同 |
| n | int | 可选 | 1 | 生成图片数量，[1,9] |
| aspect_ratio | string | 可选 | 16:9 | 生成图片纵横比（宽:高），枚举值：16:9, 9:16, 1:1, 4:3, 3:4, 3:2, 2:3, 21:9 |
| watermark_info | object | 可选 | 空 | 是否生成含水印的结果，格式：`{"enabled": boolean}` |
| callback_url | string | 可选 | 无 | 回调通知地址 |
| external_task_id | string | 可选 | 无 | 自定义任务ID，单用户下需要保证唯一性 |

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
| https | /v1/images/generations/{id} | GET |

### 请求路径参数

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| task_id | string | 必填 | 图片生成的任务ID |
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
        "watermark_info": {
            "enabled": boolean
        },
        "task_result": {
            "images": [
                {
                    "index": "int",
                    "url": "string",            // 生成图片URL（30天后清理）
                    "watermark_url": "string"   // 含水印图片URL
                }
            ]
        },
        "task_info": {
            "external_task_id": "string"
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
| https | /v1/images/generations | GET |

### 查询参数

```
/v1/images/generations?pageNum=1&pageSize=30
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
            "watermark_info": {
                "enabled": boolean
            },
            "task_result": {
                "images": [
                    {
                        "index": "int",
                        "url": "string",
                        "watermark_url": "string"
                    }
                ]
            },
            "task_info": {
                "external_task_id": "string"
            },
            "created_at": 1722769557708,
            "updated_at": 1722769557708
        }
    ]
}
```
