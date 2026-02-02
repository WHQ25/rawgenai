# 智能补全主体图

## 创建任务

| 网络协议 | 请求地址 | 请求方法 | 请求格式 | 响应格式 |
|---------|---------|---------|---------|---------|
| https | /v1/general/ai-multi-shot | POST | application/json | application/json |

## 请求头

| 字段 | 值 | 描述 |
|------|-----|------|
| Content-Type | application/json | 数据交换格式 |
| Authorization | 鉴权信息 | 鉴权信息 |

## 请求体

| 字段 | 类型 | 必填 | 默认值 | 描述 |
|------|------|------|--------|------|
| element_frontal_image | string | 是 | 无 | 主体正面参考图（Base64编码或URL），.jpg/.jpeg/.png，≤10MB，≥300px，宽高比1:2.5~2.5:1 |
| callback_url | string | 可选 | 空 | 回调通知地址 |
| external_task_id | string | 可选 | 空 | 自定义任务ID，单用户下需要保证唯一性 |

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
| https | /v1/general/ai-multi-shot/{id} | GET |

### 请求路径参数

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| task_id | string | 必填 | 任务ID |

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
        "task_result": {
            "images": [
                {
                    "index": "int",
                    "url": "string"    // 生成图片URL（30天后清理）
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
| https | /v1/general/ai-multi-shot | GET |

### 查询参数

```
/v1/general/ai-multi-shot?pageNum=1&pageSize=30
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
                "images": [
                    {
                        "index": "int",
                        "url_1": "string",   // 生成图片URL（防盗链格式，30天后清理）
                        "url_2": "string",   // 生成图片URL（防盗链格式，30天后清理）
                        "url_3": "string"    // 生成图片URL（防盗链格式，30天后清理）
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
