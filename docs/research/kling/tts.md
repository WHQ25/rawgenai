# 语音合成

## 创建任务

| 网络协议 | 请求地址 | 请求方法 | 请求格式 | 响应格式 |
|---------|---------|---------|---------|---------|
| https | /v1/audio/tts | POST | application/json | application/json |

## 请求头

| 字段 | 值 | 描述 |
|------|-----|------|
| Content-Type | application/json | 数据交换格式 |
| Authorization | 鉴权信息 | 鉴权信息 |

## 请求体

| 字段 | 类型 | 必填 | 默认值 | 描述 |
|------|------|------|--------|------|
| text | string | 是 | 无 | 合成音频的文案，最大长度1000字符 |
| voice_id | string | 是 | 无 | 音色ID，系统提供多种音色可供选择 |
| voice_language | string | 是 | zh | 音色语种，枚举值：zh（中文）、en（英文），与voice_id对应 |
| voice_speed | float | 可选 | 1.0 | 语速，有效范围0.8~2.0，精确至小数点后1位 |

## 响应体

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
            "audios": [
                {
                    "id": "string",
                    "url": "string",        // 生成音频的URL（30天后清理）
                    "duration": "string"    // 音频时长(s)
                }
            ]
        },
        "created_at": 1722769557708,
        "updated_at": 1722769557708
    }
}
```
