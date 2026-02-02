# 扩图

## 创建任务

| 网络协议 | 请求地址 | 请求方法 | 请求格式 | 响应格式 |
|---------|---------|---------|---------|---------|
| https | /v1/images/editing/expand | POST | application/json | application/json |

## 请求头

| 字段 | 值 | 描述 |
|------|-----|------|
| Content-Type | application/json | 数据交换格式 |
| Authorization | 鉴权信息 | 鉴权信息 |

## 请求体

| 字段 | 类型 | 必填 | 默认值 | 描述 |
|------|------|------|--------|------|
| image | string | 是 | 空 | 参考图片（Base64编码或URL），.jpg/.jpeg/.png，≤10MB，≥300px，宽高比1:2.5~2.5:1 |
| up_expansion_ratio | float | 是 | 0 | 向上扩充范围，基于原图高度的倍数，[0,2]，新图整体面积≤原图3倍 |
| down_expansion_ratio | float | 是 | 0 | 向下扩充范围，基于原图高度的倍数，[0,2]，新图整体面积≤原图3倍 |
| left_expansion_ratio | float | 是 | 0 | 向左扩充范围，基于原图宽度的倍数，[0,2]，新图整体面积≤原图3倍 |
| right_expansion_ratio | float | 是 | 0 | 向右扩充范围，基于原图宽度的倍数，[0,2]，新图整体面积≤原图3倍 |
| prompt | string | 可选 | 无 | 正向文本提示词，不超过2500字符 |
| n | int | 可选 | 1 | 生成图片数量，[1,9] |
| watermark_info | object | 可选 | 空 | 是否生成含水印的结果，格式：`{"enabled": boolean}` |
| callback_url | string | 可选 | 空 | 回调通知地址 |
| external_task_id | string | 可选 | 空 | 自定义任务ID，单用户下需要保证唯一性 |

**注意**：
- 扩充比例计算示例：如原图高为20，up_expansion_ratio为0.1，则原图顶边距离新图顶边为20×0.1=2
- 新图片整体面积不得超过原图片3倍

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
| https | /v1/images/editing/expand/{id} | GET |

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
        "task_info": {
            "external_task_id": "string"
        },
        "final_unit_deduction": "string",
        "watermark_info": {
            "enabled": boolean
        },
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
| https | /v1/images/editing/expand | GET |

### 查询参数

```
/v1/images/editing/expand?pageNum=1&pageSize=30
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
            "final_unit_deduction": "string",
            "watermark_info": {
                "enabled": boolean
            },
            "task_result": {
                "images": [
                    {
                        "index": "int",
                        "url": "string"
                    }
                ]
            },
            "created_at": 1722769557708,
            "updated_at": 1722769557708
        }
    ]
}
```
