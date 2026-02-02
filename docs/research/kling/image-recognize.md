# 图像识别

## 创建任务

| 网络协议 | 请求地址 | 请求方法 | 请求格式 | 响应格式 |
|---------|---------|---------|---------|---------|
| https | /v1/videos/image-recognize | POST | application/json | application/json |

## 请求头

| 字段 | 值 | 描述 |
|------|-----|------|
| Content-Type | application/json | 数据交换格式 |
| Authorization | 鉴权信息 | 鉴权信息 |

## 请求体

| 字段 | 类型 | 必填 | 默认值 | 描述 |
|------|------|------|--------|------|
| image | string | 是 | 无 | 待识别的图片（Base64编码或URL），.jpg/.jpeg/.png，≤10MB，≥300px，宽高比1:2.5~2.5:1 |

## 响应体

```json
{
    "code": 0,
    "message": "string",
    "request_id": "string",
    "data": {
        "final_unit_deduction": "string",
        "task_result": {
            "images": [
                {
                    "type": "object_seg",
                    "is_contain": true,
                    "url": "string"    // 识别后图片URL（30天后清理）
                },
                {
                    "type": "head_seg",
                    "is_contain": true,
                    "url": "string"    // 含头发的人物面部识别结果（30天后清理）
                },
                {
                    "type": "face_seg",
                    "is_contain": true,
                    "url": "string"    // 不含头发的人物面部识别结果（30天后清理）
                },
                {
                    "type": "cloth_seg",
                    "is_contain": true,
                    "url": "string"    // 服装识别结果（30天后清理）
                }
            ]
        }
    }
}
```

**识别类型说明**：
- `object_seg` - 主体识别，识别图片中的主要对象
- `head_seg` - 含头发的人物面部识别
- `face_seg` - 不含头发的人物面部识别
- `cloth_seg` - 服装识别

**注意**：
- `is_contain` 为 true 表示识别到对应的内容，false 表示未识别到
- 每种识别类型都会返回对应的分割结果图片
