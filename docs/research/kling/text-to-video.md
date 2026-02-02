# 文生视频

## 创建任务

| 网络协议 | 请求地址 | 请求方法 | 请求格式 | 响应格式 |
|---------|---------|---------|---------|---------|
| https | /v1/videos/text2video | POST | application/json | application/json |

## 请求头

| 字段 | 值 | 描述 |
|------|-----|------|
| Content-Type | application/json | 数据交换格式 |
| Authorization | 鉴权信息，参考接口鉴权 | 鉴权信息 |

## 请求体

| 字段 | 类型 | 必填 | 默认值 | 描述 |
|------|------|------|--------|------|
| model_name | string | 可选 | kling-v1 | 模型名称。枚举值：`kling-v1`, `kling-v1-6`, `kling-v2-master`, `kling-v2-1-master`, `kling-v2-5-turbo`, `kling-v2-6` |
| prompt | string | 必须 | 无 | 正向文本提示词，不能超过2500个字符 |
| negative_prompt | string | 可选 | 空 | 负向文本提示词，不能超过2500个字符 |
| sound | string | 可选 | off | 生成视频时是否同时生成声音。枚举值：`on`, `off`。仅V2.6及后续版本模型支持 |
| cfg_scale | float | 可选 | 0.5 | 生成视频的自由度。值越大，与提示词相关性越强。取值范围：[0, 1]。kling-v2.x模型不支持 |
| mode | string | 可选 | std | 生成视频的模式。`std`：标准模式，`pro`：专家模式（高品质） |
| camera_control | object | 可选 | 空 | 控制摄像机运动的协议 |
| aspect_ratio | string | 可选 | 16:9 | 画面纵横比。枚举值：`16:9`, `9:16`, `1:1` |
| duration | string | 可选 | 5 | 生成视频时长（秒）。枚举值：`5`, `10` |
| watermark_info | array | 可选 | 空 | 是否生成含水印的结果 |
| callback_url | string | 可选 | 无 | 任务结果回调通知地址 |
| external_task_id | string | 可选 | 无 | 自定义任务ID |

### camera_control 运镜控制

#### type - 预定义的运镜类型

| 值 | 描述 | config |
|----|------|--------|
| simple | 简单运镜，可在config中六选一 | 必填 |
| down_back | 镜头下压并后退（下移拉远） | 不填 |
| forward_up | 镜头前进并上仰（推进上移） | 不填 |
| right_turn_forward | 先右旋转后前进（右旋推进） | 不填 |
| left_turn_forward | 先左旋并前进（左旋推进） | 不填 |

#### config - 运镜参数（type=simple时必填，6选1）

| 参数 | 描述 | 取值范围 |
|------|------|----------|
| horizontal | 水平运镜，控制摄像机沿x轴平移 | [-10, 10]，负值向左，正值向右 |
| vertical | 垂直运镜，控制摄像机沿y轴平移 | [-10, 10]，负值向下，正值向上 |
| pan | 水平摇镜，控制摄像机绕y轴旋转 | [-10, 10]，负值向左旋转，正值向右旋转 |
| tilt | 垂直摇镜，控制摄像机绕x轴旋转 | [-10, 10]，负值向下旋转，正值向上旋转 |
| roll | 旋转运镜，控制摄像机绕z轴旋转 | [-10, 10]，负值逆时针，正值顺时针 |
| zoom | 变焦，控制摄像机的焦距变化 | [-10, 10]，负值焦距变长（视野变小），正值焦距变短（视野变大） |

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
    "request_id": "string",             // 请求ID
    "data": {
        "task_id": "string",            // 任务ID，系统生成
        "task_info": {
            "external_task_id": "string" // 客户自定义任务ID
        },
        "task_status": "string",        // submitted/processing/succeed/failed
        "created_at": 1722769557708,    // 创建时间(ms)
        "updated_at": 1722769557708     // 更新时间(ms)
    }
}
```

---

## 查询任务（单个）

| 网络协议 | 请求地址 | 请求方法 | 请求格式 | 响应格式 |
|---------|---------|---------|---------|---------|
| https | /v1/videos/text2video/{id} | GET | application/json | application/json |

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
        "task_status": "string",
        "task_status_msg": "string",        // 失败原因
        "task_info": {
            "external_task_id": "string"
        },
        "task_result": {
            "videos": [
                {
                    "id": "string",
                    "url": "string",            // 视频URL（30天后清理）
                    "watermark_url": "string",
                    "duration": "string"
                }
            ]
        },
        "watermark_info": {
            "enabled": boolean
        },
        "final_unit_deduction": "string",   // 扣减积分数值
        "created_at": 1722769557708,
        "updated_at": 1722769557708
    }
}
```

---

## 查询任务（列表）

| 网络协议 | 请求地址 | 请求方法 | 请求格式 | 响应格式 |
|---------|---------|---------|---------|---------|
| https | /v1/videos/text2video | GET | application/json | application/json |

### 查询参数

```
/v1/videos/text2video?pageNum=1&pageSize=30
```

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
