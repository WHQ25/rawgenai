# Omni-Video (O1)

## 创建任务

| 网络协议 | 请求地址 | 请求方法 | 请求格式 | 响应格式 |
|---------|---------|---------|---------|---------|
| https | /v1/videos/omni-video | POST | application/json | application/json |

## 请求头

| 字段 | 值 | 描述 |
|------|-----|------|
| Content-Type | application/json | 数据交换格式 |
| Authorization | 鉴权信息，参考接口鉴权 | 鉴权信息，参考接口鉴权 |

## 请求体

| 字段 | 类型 | 必填 | 默认值 | 描述 |
|------|------|------|--------|------|
| model_name | string | 可选 | kling-v1 | 模型名称，枚举值：`kling-video-o1` |
| prompt | string | 必须 | 无 | 文本提示词，可包含正向描述和负向描述。可将提示词模板化来满足不同的视频生成需求。不能超过2500个字符。Omni模型可通过Prompt与主体、图片、视频等内容实现多种能力，通过`<<<>>>`的格式来指定某个主体、图片或视频，如：`<<<element_1>>>`、`<<<image_1>>>`、`<<<video_1>>>` |
| image_list | array | 可选 | 空 | 参考图列表。包括主体、场景、风格等参考图片，也可作为首帧或尾帧生成视频 |
| element_list | array | 可选 | 空 | 主体参考列表，基于主体库中主体的ID配置 |
| video_list | array | 可选 | 空 | 参考视频，通过URL方式获取。可作为特征参考视频，也可作为待编辑视频 |
| mode | string | 可选 | pro | 生成视频的模式。`std`：标准模式（性价比高），`pro`：专家模式（高品质） |
| aspect_ratio | string | 可选 | 空 | 生成视频的画面纵横比（宽:高）。枚举值：`16:9`, `9:16`, `1:1`。未使用首帧参考或视频编辑功能时必填 |
| duration | string | 可选 | 5 | 生成视频时长（秒）。枚举值：`3-10`。文生视频、首帧图生视频仅支持5和10s |
| watermark_info | array | 可选 | 空 | 是否同时生成含水印的结果 |
| callback_url | string | 可选 | 空 | 本次任务结果回调通知地址 |
| external_task_id | string | 可选 | 空 | 自定义任务ID |

### image_list 结构

```json
"image_list": [
    {
        "image_url": "image_url",
        "type": "first_frame"  // first_frame=首帧, end_frame=尾帧
    },
    {
        "image_url": "image_url",
        "type": "end_frame"
    }
]
```

**图片限制**:
- 支持 Base64 编码或 URL（确保可访问）
- 格式支持：.jpg / .jpeg / .png
- 文件大小：≤10MB
- 宽高尺寸：≥300px
- 宽高比：1:2.5 ~ 2.5:1
- 有参考视频时：≤4张；无参考视频时：≤7张
- 超过2张图片时，不支持设置尾帧

### element_list 结构

```json
"element_list": [
    {
        "element_id": long
    }
]
```

**主体限制**:
- 有参考视频时：参考图片数量 + 参考主体数量 ≤ 4
- 无参考视频时：参考图片数量 + 参考主体数量 ≤ 7

### video_list 结构

```json
"video_list": [
    {
        "video_url": "video_url",
        "refer_type": "base",  // feature=特征参考视频, base=待编辑视频
        "keep_original_sound": "yes"  // yes=保留原声, no=不保留
    }
]
```

**视频限制**:
- 格式：仅支持 MP4/MOV
- 时长：3-10秒
- 宽高尺寸：720px - 2160px
- 帧率：24fps - 60fps（生成视频输出为24fps）
- 数量：≤1段
- 大小：≤200MB

### watermark_info 结构

```json
"watermark_info": {
    "enabled": boolean  // true=生成水印, false=不生成
}
```

## 响应体

```json
{
    "code": 0,                          // 错误码
    "message": "string",                // 错误信息
    "request_id": "string",             // 请求ID，用于跟踪请求
    "data": {
        "task_id": "string",            // 任务ID，系统生成
        "task_info": {
            "external_task_id": "string" // 客户自定义任务ID
        },
        "task_status": "string",        // submitted/processing/succeed/failed
        "created_at": 1722769557708,    // 创建时间，Unix时间戳(ms)
        "updated_at": 1722769557708     // 更新时间，Unix时间戳(ms)
    }
}
```

## 更多场景调用示例

### 图片/主体参考

参考图片/主体里的角色/道具/场景等多种元素，灵活生成视频

```bash
curl --location 'https://api-beijing.klingai.com/v1/videos/omni-video' \
--header 'Authorization: Bearer xxx' \
--header 'Content-Type: application/json' \
--data '{
    "model_name": "kling-video-o1",
    "prompt": "<<<image_1>>>在东京的街头漫步，偶遇<<<element_1>>>和<<<element_2>>>，并跳到<<<element_2>>>的怀里。视频画面风格与<<<image_2>>>相同",
    "image_list": [
        { "image_url": "xxxxx" },
        { "image_url": "xxxxx" }
    ],
    "element_list": [
        { "element_id": long },
        { "element_id": long }
    ],
    "mode": "pro",
    "aspect_ratio": "1:1",
    "duration": "7"
}'
```

### 指令变换

视频编辑，例如视频增加内容/删除内容/修改内容（主体/背景/局部/视频风格/物体颜色/天气/…）/切换景别/切换视角

```bash
curl --location 'https://api-beijing.klingai.com/v1/videos/omni-video' \
--header 'Authorization: Bearer xxx' \
--header 'Content-Type: application/json' \
--data '{
    "model_name": "kling-video-o1",
    "prompt": "给<<<video_1>>>中的穿蓝衣服的女孩，戴上<<<image_1>>>中的王冠",
    "image_list": [
        { "image_url": "xxx" }
    ],
    "video_list": [
        {
            "video_url": "xxxxxxxx",
            "refer_type": "base",
            "keep_original_sound": "yes"
        }
    ],
    "mode": "pro"
}'
```

### 视频参考

参考视频内容生成下一个镜头/上一个镜头，或者参考视频的风格/运镜方式进行视频生成

```bash
curl --location 'https://api-beijing.klingai.com/v1/videos/omni-video' \
--header 'Authorization: Bearer xxx' \
--header 'Content-Type: application/json' \
--data '{
    "model_name": "kling-video-o1",
    "prompt": "参考<<<video_1>>>的运镜方式，生成一段视频：<<<element_1>>>和<<<element_2>>>在东京街头漫步，偶遇<<<image_1>>>",
    "image_list": [
        { "image_url": "xxx" }
    ],
    "element_list": [
        { "element_id": "xxx" },
        { "element_id": "xxx" }
    ],
    "video_list": [
        {
            "video_url": "xxxxxxxx",
            "refer_type": "feature",
            "keep_original_sound": "yes"
        }
    ],
    "mode": "pro",
    "aspect_ratio": "1:1",
    "duration": "7"
}'
```

### 视频延长（生成下一个镜头）

```bash
curl --location 'https://api-beijing.klingai.com/v1/videos/omni-video' \
--header 'Authorization: Bearer xxx' \
--header 'Content-Type: application/json' \
--data '{
    "model_name": "kling-video-o1",
    "prompt": "基于<<<video_1>>>，生成下一个镜头",
    "video_list": [
        {
            "video_url": "xxxxxxxx",
            "refer_type": "feature",
            "keep_original_sound": "yes"
        }
    ],
    "mode": "pro"
}'
```

### 首尾帧图生视频

```bash
curl --location 'https://api-beijing.klingai.com/v1/videos/omni-video' \
--header 'Authorization: Bearer xxx' \
--header 'Content-Type: application/json' \
--data '{
    "model_name": "kling-video-o1",
    "prompt": "视频中的人跳舞",
    "image_list": [
        {
            "image_url": "xxx",
            "type": "first_frame"
        },
        {
            "image_url": "xxx",
            "type": "end_frame"
        }
    ],
    "mode": "pro"
}'
```

### 文生视频

```bash
curl --location 'https://api-beijing.klingai.com/v1/videos/omni-video' \
--header 'Authorization: Bearer xxx' \
--header 'Content-Type: application/json' \
--data '{
    "model_name": "kling-video-o1",
    "prompt": "视频中的人跳舞",
    "mode": "pro",
    "aspect_ratio": "1:1",
    "duration": "7"
}'
```

## FAQ

### 1. duration 什么情况支持？

- **文生/图生（不含首尾帧）**：可选 5s/10s
- **有视频输入且使用视频编辑功能（类型=base）**：不可指定时长，跟视频对齐
- **其他情况**：可选 3-10s

### 2. 怎么进行视频延长？

通过"视频参考"来实现，传入一段视频，通过prompt驱动模型"生成下一个镜头"或者"生成上一个镜头"

### 3. aspect_ratio 什么情况支持？

- **不支持**：指令变换（视频编辑），图生视频（包括首尾帧）
- **支持**：文生视频，图片/主体参考，视频参考-其他，视频参考-生成下一个/上一个镜头

---

## 查询任务（单个）

| 网络协议 | 请求地址 | 请求方法 | 请求格式 | 响应格式 |
|---------|---------|---------|---------|---------|
| https | /v1/videos/omni-video/{id} | GET | application/json | application/json |

### 请求路径参数

| 字段 | 类型 | 必填 | 默认值 | 描述 |
|------|------|------|--------|------|
| task_id | string | 可选 | 无 | 任务ID（路径参数），与 external_task_id 二选一 |
| external_task_id | string | 可选 | 无 | 自定义任务ID，与 task_id 二选一 |

### 响应体

```json
{
    "code": 0,
    "message": "string",
    "request_id": "string",
    "data": {
        "task_id": "string",
        "task_status": "string",            // submitted/processing/succeed/failed
        "task_status_msg": "string",        // 失败原因（如触发内容风控）
        "task_info": {
            "external_task_id": "string"
        },
        "task_result": {
            "videos": [
                {
                    "id": "string",         // 生成的视频ID，全局唯一
                    "url": "string",        // 生成视频的URL（30天后清理）
                    "watermark_url": "string", // 含水印视频URL
                    "duration": "string"    // 视频总时长(s)
                }
            ]
        },
        "watermark_info": {
            "enabled": boolean
        },
        "final_unit_deduction": "string",   // 任务最终扣减积分数值
        "created_at": 1722769557708,
        "updated_at": 1722769557708
    }
}
```

---

## 查询任务（列表）

| 网络协议 | 请求地址 | 请求方法 | 请求格式 | 响应格式 |
|---------|---------|---------|---------|---------|
| https | /v1/videos/omni-video | GET | application/json | application/json |

### 查询参数

| 字段 | 类型 | 必填 | 默认值 | 描述 |
|------|------|------|--------|------|
| pageNum | int | 可选 | 1 | 页码，取值范围：[1,1000] |
| pageSize | int | 可选 | 30 | 每页数据量，取值范围：[1,500] |

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
