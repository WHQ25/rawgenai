# 多模态视频编辑

## 概述

多模态视频编辑功能需要先对原始视频进行初始化处理，在替换或删除现有视频中的元素时，需先标记视频中相关元素。

---

## 初始化待编辑视频

| 网络协议 | 请求地址 | 请求方法 | 请求格式 | 响应格式 |
|---------|---------|---------|---------|---------|
| https | /v1/videos/multi-elements/init-selection | POST | application/json | application/json |

### 请求头

| 字段 | 值 | 描述 |
|------|-----|------|
| Content-Type | application/json | 数据交换格式 |
| Authorization | 鉴权信息 | 鉴权信息 |

### 请求体

| 字段 | 类型 | 必填 | 默认值 | 描述 |
|------|------|------|--------|------|
| video_id | string | 可选 | 空 | 视频ID，仅支持30天内生成的作品。与 video_url 二选一 |
| video_url | string | 可选 | 无 | 视频URL。与 video_id 二选一 |

**视频限制**：
- 时长：2-5秒 或 7-10秒
- 格式：MP4 或 MOV
- 宽高尺寸：720px - 2160px
- 帧率：24、30 或 60fps

### 响应体

```json
{
    "code": 0,
    "message": "string",
    "request_id": "string",
    "data": {
        "status": 0,                    // 拒识码，非0为失败
        "session_id": "id",             // 会话ID，有效期24小时
        "final_unit_deduction": "string",
        "fps": 30.0,                    // 解析后的帧数
        "original_duration": 1000,      // 视频时长(ms)
        "width": 720,
        "height": 1280,
        "total_frame": 300,             // 总帧数
        "normalized_video": "url"       // 初始化后的视频URL
    }
}
```

---

## 增加视频选区

| 网络协议 | 请求地址 | 请求方法 |
|---------|---------|---------|
| https | /v1/videos/multi-elements/add-selection | POST |

### 请求体

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| session_id | string | 是 | 会话ID |
| frame_index | int | 是 | 帧号 |
| points | object[] | 是 | 点选坐标数组 |

**points 结构**：
```json
{
    "x": 0.5,  // [0,1] 百分比，0代表左边，1代表右边
    "y": 0.5   // [0,1] 百分比，0代表上面，1代表下面
}
```

### 响应体

```json
{
    "code": 0,
    "message": "string",
    "request_id": "string",
    "data": {
        "status": 0,
        "session_id": "id",
        "final_unit_deduction": "string",
        "res": {
            "frame_index": 0,
            "rle_mask_list": [
                {
                    "object_id": 0,
                    "rle_mask": {
                        "size": [720, 1280],
                        "counts": "string"
                    },
                    "png_mask": {
                        "size": [720, 1280],
                        "base64": "string"
                    }
                }
            ]
        }
    }
}
```

---

## 删减视频选区

| 网络协议 | 请求地址 | 请求方法 |
|---------|---------|---------|
| https | /v1/videos/multi-elements/delete-selection | POST |

### 请求体

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| session_id | string | 是 | 会话ID |
| frame_index | int | 是 | 帧号 |
| points | object[] | 是 | 点选坐标数组（需与增加时完全一致） |

---

## 清除视频选区

| 网络协议 | 请求地址 | 请求方法 |
|---------|---------|---------|
| https | /v1/videos/multi-elements/clear-selection | POST |

### 请求体

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| session_id | string | 是 | 会话ID |

---

## 预览已选区视频

| 网络协议 | 请求地址 | 请求方法 |
|---------|---------|---------|
| https | /v1/videos/multi-elements/preview-selection | POST |

### 请求体

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| session_id | string | 是 | 会话ID |

### 响应体

```json
{
    "code": 0,
    "data": {
        "status": 0,
        "session_id": "id",
        "final_unit_deduction": "string",
        "res": {
            "video": "url",              // 含mask的视频
            "video_cover": "url",        // 视频封面
            "tracking_output": "url"     // mask结果
        }
    }
}
```

---

## 创建任务

| 网络协议 | 请求地址 | 请求方法 |
|---------|---------|---------|
| https | /v1/videos/multi-elements | POST |

### 请求体

| 字段 | 类型 | 必填 | 默认值 | 描述 |
|------|------|------|--------|------|
| model_name | string | 可选 | kling-v1-6 | 模型名称 |
| session_id | string | 是 | 无 | 会话ID |
| edit_mode | string | 是 | 无 | 操作类型：`addition`(增加), `swap`(替换), `removal`(删除) |
| image_list | array | 可选 | 空 | 参考图像列表 |
| prompt | string | 是 | 无 | 提示词，不能超过2500字符 |
| negative_prompt | string | 可选 | 空 | 负向提示词 |
| mode | string | 可选 | std | 生成模式：`std`, `pro` |
| duration | string | 可选 | 5 | 视频时长：`5`, `10` |
| watermark_info | array | 可选 | 空 | 水印配置 |
| callback_url | string | 可选 | 空 | 回调地址 |
| external_task_id | string | 可选 | 空 | 自定义任务ID |

### image_list 结构

```json
"image_list": [
    { "image": "image_url_1" },
    { "image": "image_url_2" }
]
```

**图片限制**：
- addition（增加）：1-2张
- swap（替换）：1张
- removal（删除）：无需填写
- 格式：.jpg / .jpeg / .png
- 大小：≤10MB
- 尺寸：≥300px
- 宽高比：1:2.5 ~ 2.5:1

### Prompt 模板

#### 增加元素

中文：`基于<<<video_1>>>中的原始内容，以自然生动的方式，将<<<image_1>>>中的【】，融入<<<video_1>>>的【】`

英文：`Using the context of <<<video_1>>>, seamlessly add [x] from <<<image_1>>>`

#### 替换元素

中文：`使用<<<image_1>>>中的【】，替换<<<video_1>>>中的【】`

英文：`swap [x] from <<<image_1>>> for [x] from <<<video_1>>>`

#### 删除元素

中文：`删除<<<video_1>>>中的【】`

英文：`Delete [x] from <<<video_1>>>`

### 时长限制

- 生成5s视频：输入视频需2-5s
- 生成10s视频：输入视频需7-10s

---

## 查询任务（单个）

| 网络协议 | 请求地址 | 请求方法 |
|---------|---------|---------|
| https | /v1/videos/multi-elements/{id} | GET |

### 请求路径参数

| 字段 | 类型 | 描述 |
|------|------|------|
| task_id | string | 任务ID（与 external_task_id 二选一） |
| external_task_id | string | 自定义任务ID |

### 响应体

```json
{
    "code": 0,
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
                    "session_id": "id",
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
| https | /v1/videos/multi-elements | GET |

### 查询参数

```
/v1/videos/multi-elements?pageNum=1&pageSize=30
```

| 字段 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| pageNum | int | 1 | 页码 |
| pageSize | int | 30 | 每页数据量 |
