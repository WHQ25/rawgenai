# 多图参考生视频

## 创建任务

| 网络协议 | 请求地址 | 请求方法 | 请求格式 | 响应格式 |
|---------|---------|---------|---------|---------|
| https | /v1/videos/multi-image2video | POST | application/json | application/json |

## 请求头

| 字段 | 值 | 描述 |
|------|-----|------|
| Content-Type | application/json | 数据交换格式 |
| Authorization | 鉴权信息，参考接口鉴权 | 鉴权信息 |

## 请求体

| 字段 | 类型 | 必填 | 默认值 | 描述 |
|------|------|------|--------|------|
| model_name | string | 可选 | kling-v1-6 | 模型名称。枚举值：`kling-v1-6` |
| image_list | array | 必须 | 空 | 参考图片列表，最多4张 |
| prompt | string | 必须 | 无 | 正向文本提示词，不能超过2500个字符 |
| negative_prompt | string | 可选 | 空 | 负向文本提示词，不能超过2500个字符 |
| mode | string | 可选 | std | 生成模式。`std`：标准模式，`pro`：专家模式 |
| duration | string | 可选 | 5 | 视频时长（秒）。枚举值：`5`, `10` |
| aspect_ratio | string | 可选 | 16:9 | 画面纵横比。枚举值：`16:9`, `9:16`, `1:1` |
| watermark_info | array | 可选 | 空 | 水印配置 |
| callback_url | string | 可选 | 无 | 回调通知地址 |
| external_task_id | string | 可选 | 无 | 自定义任务ID |

### image_list 结构

```json
"image_list": [
    { "image": "image_url_1" },
    { "image": "image_url_2" },
    { "image": "image_url_3" },
    { "image": "image_url_4" }
]
```

**图片限制**：
- 最多4张图片
- API端无裁剪逻辑，请直接上传已选主体后的图片
- 支持 Base64 编码或 URL（确保可访问）
- **Base64格式**：不要添加 `data:image/png;base64,` 前缀
- 格式支持：.jpg / .jpeg / .png
- 文件大小：≤10MB
- 宽高尺寸：≥300px
- 宽高比：1:2.5 ~ 2.5:1

---

## 查询任务（单个）

| 网络协议 | 请求地址 | 请求方法 |
|---------|---------|---------|
| https | /v1/videos/multi-image2video/{id} | GET |

### 请求路径参数

| 字段 | 类型 | 描述 |
|------|------|------|
| task_id | string | 任务ID（路径参数），与 external_task_id 二选一 |
| external_task_id | string | 自定义任务ID，与 task_id 二选一 |

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
        "final_unit_deduction": "string",
        "watermark_info": {
            "enabled": boolean
        },
        "task_result": {
            "videos": [
                {
                    "id": "string",
                    "url": "string",
                    "duration": "string"
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
| https | /v1/videos/multi-image2video | GET |

### 查询参数

```
/v1/videos/multi-image2video?pageNum=1&pageSize=30
```

| 字段 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| pageNum | int | 1 | 页码，[1,1000] |
| pageSize | int | 30 | 每页数据量，[1,500] |
