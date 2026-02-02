# 动作控制

## 创建任务

| 网络协议 | 请求地址 | 请求方法 | 请求格式 | 响应格式 |
|---------|---------|---------|---------|---------|
| https | /v1/videos/motion-control | POST | application/json | application/json |

## 请求头

| 字段 | 值 | 描述 |
|------|-----|------|
| Content-Type | application/json | 数据交换格式 |
| Authorization | 鉴权信息，参考接口鉴权 | 鉴权信息 |

## 请求体

| 字段 | 类型 | 必填 | 默认值 | 描述 |
|------|------|------|--------|------|
| prompt | string | 可选 | 空 | 文本提示词，可通过提示词为画面增加元素、实现运镜效果等，不能超过2500个字符 |
| image_url | string | 必须 | 无 | 参考图像，生成视频中的人物、背景等元素均以参考图为准 |
| video_url | string | 必须 | 无 | 参考视频，生成视频中的人物动作与参考视频一致 |
| keep_original_sound | string | 可选 | yes | 是否保留视频原声。枚举值：`yes`, `no` |
| character_orientation | string | 必须 | 无 | 生成视频中人物的朝向。枚举值：`image`（与图片一致）, `video`（与视频一致） |
| mode | string | 必须 | 无 | 生成模式。`std`：标准模式，`pro`：专家模式 |
| watermark_info | array | 可选 | 空 | 水印配置 |
| callback_url | string | 可选 | 无 | 回调通知地址 |
| external_task_id | string | 可选 | 无 | 自定义任务ID |

### 参考图像要求 (image_url)

- 人物比例尽量与参考动作比例一致
- 人物需要露出清晰的上半身或全身的肢体及头部，避免遮挡
- 避免存在极端朝向（倒立、平卧等）
- 人物占画面比例不得太低
- 支持真实/风格化的角色（人物/类人动物/部分纯动物/部分类人肢体比例的角色）
- 支持 Base64 编码或 URL
- 格式：.jpg / .jpeg / .png
- 文件大小：≤10MB
- 宽高尺寸：300px ~ 65536px
- 宽高比：1:2.5 ~ 2.5:1

### 参考视频要求 (video_url)

- 人物需要露出清晰的上半身或全身的全部肢体及头部，避免遮挡
- 建议上传1人动作视频（2人及以上会取画面占比最大的人物动作）
- 推荐使用真人动作
- 视频一镜到底，避免切镜、运镜
- 动作避免过快，相对平稳的动作效果更佳
- 格式：.mp4 / .mov
- 文件大小：≤100MB
- 边长：340px ~ 3850px
- **时长限制**：
  - 最短：3秒
  - `character_orientation=video` 时：最长30秒
  - `character_orientation=image` 时：最长10秒

### character_orientation 人物朝向

| 值 | 描述 | 视频时长限制 |
|----|------|-------------|
| image | 与图片中人物朝向一致 | ≤10秒 |
| video | 与视频中人物朝向一致 | ≤30秒 |

---

## 查询任务（单个）

| 网络协议 | 请求地址 | 请求方法 |
|---------|---------|---------|
| https | /v1/videos/motion-control/{id} | GET |

### 请求路径参数

| 字段 | 类型 | 描述 |
|------|------|------|
| task_id | string | 任务ID（路径参数） |
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
}
```

---

## 查询任务（列表）

| 网络协议 | 请求地址 | 请求方法 |
|---------|---------|---------|
| https | /v1/videos/motion-control | GET |

### 查询参数

```
/v1/videos/motion-control?pageNum=1&pageSize=30
```

| 字段 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| pageNum | int | 1 | 页码，[1,1000] |
| pageSize | int | 30 | 每页数据量，[1,500] |
