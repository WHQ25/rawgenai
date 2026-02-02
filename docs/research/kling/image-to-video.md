# 图生视频

## 创建任务

| 网络协议 | 请求地址 | 请求方法 | 请求格式 | 响应格式 |
|---------|---------|---------|---------|---------|
| https | /v1/videos/image2video | POST | application/json | application/json |

## 请求头

| 字段 | 值 | 描述 |
|------|-----|------|
| Content-Type | application/json | 数据交换格式 |
| Authorization | 鉴权信息，参考接口鉴权 | 鉴权信息 |

## 请求体

| 字段 | 类型 | 必填 | 默认值 | 描述 |
|------|------|------|--------|------|
| model_name | string | 可选 | kling-v1 | 模型名称。枚举值：`kling-v1`, `kling-v1-5`, `kling-v1-6`, `kling-v2-master`, `kling-v2-1`, `kling-v2-1-master`, `kling-v2-5-turbo`, `kling-v2-6` |
| image | string | 必须 | 空 | 参考图像（首帧），支持 Base64 或 URL |
| image_tail | string | 可选 | 空 | 尾帧图像，支持 Base64 或 URL |
| prompt | string | 可选 | 无 | 正向文本提示词，不能超过2500个字符 |
| negative_prompt | string | 可选 | 空 | 负向文本提示词，不能超过2500个字符 |
| voice_list | array | 可选 | 无 | 音色列表，最多2个。仅V2.6及后续版本支持 |
| sound | string | 可选 | off | 是否生成声音。枚举值：`on`, `off`。仅V2.6及后续版本支持 |
| cfg_scale | float | 可选 | 0.5 | 生成视频的自由度。取值范围：[0, 1]。kling-v2.x不支持 |
| mode | string | 可选 | std | 生成模式。`std`：标准模式，`pro`：专家模式 |
| static_mask | string | 可选 | 无 | 静态笔刷涂抹区域 |
| dynamic_masks | array | 可选 | 无 | 动态笔刷配置列表，最多6组 |
| camera_control | object | 可选 | 空 | 运镜控制 |
| duration | string | 可选 | 5 | 视频时长（秒）。枚举值：`5`, `10` |
| watermark_info | array | 可选 | 空 | 水印配置 |
| callback_url | string | 可选 | 无 | 回调通知地址 |
| external_task_id | string | 可选 | 无 | 自定义任务ID |

### 图片限制

- 支持 Base64 编码或 URL（确保可访问）
- **Base64格式**：不要添加 `data:image/png;base64,` 前缀
- 格式支持：.jpg / .jpeg / .png
- 文件大小：≤10MB
- 宽高尺寸：≥300px
- 宽高比：1:2.5 ~ 2.5:1

### 互斥参数

以下三组参数三选一，不能同时使用：
1. `image` + `image_tail`（首尾帧）
2. `dynamic_masks` / `static_mask`（运动笔刷）
3. `camera_control`（运镜控制）

### 运动笔刷

#### static_mask - 静态笔刷

静态笔刷涂抹区域，图片长宽比必须与输入图片相同。

#### dynamic_masks - 动态笔刷

| 字段 | 类型 | 描述 |
|------|------|------|
| mask | string | 动态笔刷涂抹区域图片 |
| trajectories | array | 运动轨迹坐标序列 |

**trajectories 说明**：
- 坐标个数：[2, 77]
- 坐标系：以图片**左下角**为原点
- 每个坐标点：`{"x": int, "y": int}`
- 坐标点越多轨迹越准确
- 轨迹方向以传入顺序为指向

### camera_control 运镜控制

同文生视频的 camera_control 参数。

#### type - 预定义的运镜类型

| 值 | 描述 |
|----|------|
| simple | 简单运镜，需配置 config |
| down_back | 下移拉远 |
| forward_up | 推进上移 |
| right_turn_forward | 右旋推进 |
| left_turn_forward | 左旋推进 |

#### config 参数（type=simple时必填，6选1）

| 参数 | 描述 | 取值范围 |
|------|------|----------|
| horizontal | 水平运镜（x轴平移） | [-10, 10] |
| vertical | 垂直运镜（y轴平移） | [-10, 10] |
| pan | 水平摇镜（绕y轴） | [-10, 10] |
| tilt | 垂直摇镜（绕x轴） | [-10, 10] |
| roll | 旋转运镜（绕z轴） | [-10, 10] |
| zoom | 变焦 | [-10, 10] |

### voice_list 音色配置（V2.6+）

```json
"voice_list": [
    {"voice_id": "voice_id_1"},
    {"voice_id": "voice_id_2"}
]
```

- 在 prompt 中使用 `<<<voice_1>>>` 指定音色
- 最多引用2个音色
- 指定音色时 sound 必须为 `on`

## 示例请求

```bash
curl --location --request POST 'https://api-beijing.klingai.com/v1/videos/image2video' \
--header 'Authorization: Bearer xxx' \
--header 'Content-Type: application/json' \
--data-raw '{
    "model_name": "kling-v1",
    "mode": "pro",
    "duration": "5",
    "image": "https://example.com/image.jpg",
    "prompt": "宇航员站起身走了",
    "cfg_scale": 0.5,
    "static_mask": "https://example.com/static_mask.png",
    "dynamic_masks": [
        {
            "mask": "https://example.com/dynamic_mask.png",
            "trajectories": [
                {"x": 279, "y": 219},
                {"x": 417, "y": 65}
            ]
        }
    ]
}'
```

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
| https | /v1/videos/image2video/{id} | GET |

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
                    "url": "string",      // 30天后清理
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
| https | /v1/videos/image2video | GET |

### 查询参数

```
/v1/videos/image2video?pageNum=1&pageSize=30
```

| 字段 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| pageNum | int | 1 | 页码，[1,1000] |
| pageSize | int | 30 | 每页数据量，[1,500] |
