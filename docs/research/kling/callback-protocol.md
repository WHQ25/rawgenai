# Callback 协议

## 概述

对于异步任务（图像生成、视频生成、虚拟试穿等），若在创建任务时设置了 `callback_url`，服务端会在任务状态变更时主动通知。

## 回调通知格式

当任务状态变更时，服务端会向配置的 `callback_url` 发送 POST 请求，包含以下信息：

```json
{
    "task_id": "string",
    "task_status": "string",  // submitted/processing/succeed/failed
    "task_status_msg": "string",
    "task_info": {
        "parent_video": {
            "id": "string",
            "url": "string",      // 30天后清理
            "duration": "string"
        },
        "external_task_id": "string"
    },
    "created_at": 1722769557708,
    "updated_at": 1722769557708,
    "task_result": {
        "images": [
            {
                "index": "int",
                "url": "string"   // 30天后清理
            }
        ],
        "videos": [
            {
                "id": "string",
                "url": "string",  // 30天后清理
                "duration": "string"
            }
        ]
    }
}
```

## 字段说明

| 字段 | 类型 | 描述 |
|------|------|------|
| task_id | string | 任务ID，系统生成 |
| task_status | string | 任务状态，枚举值：submitted（已提交）、processing（处理中）、succeed（成功）、failed（失败） |
| task_status_msg | string | 任务状态信息，当任务失败时展示失败原因 |
| task_info | object | 任务创建时的参数信息 |
| external_task_id | string | 客户自定义任务ID |
| parent_video | object | 源视频信息（续写类任务） |
| created_at | long | 任务创建时间，Unix时间戳，单位ms |
| updated_at | long | 任务更新时间，Unix时间戳，单位ms |
| task_result | object | 任务结果 |
| images | array | 图片类任务的结果列表 |
| videos | array | 视频类任务的结果列表 |

## 任务状态说明

| 状态 | 描述 |
|------|------|
| submitted | 已提交，任务刚创建 |
| processing | 处理中，任务正在执行 |
| succeed | 成功，任务完成且结果可用 |
| failed | 失败，任务执行失败 |

## 重要提示

- **数据有效期**：生成的图片和视频会在 30 天后被清理，请及时转存
- **通知可靠性**：建议在接收到通知后验证 task_id，可通过查询接口二次确认任务状态
- **重试机制**：建议实现重试逻辑处理网络异常导致的通知丢失
- **幂等性**：建议根据 task_id 和 external_task_id 实现幂等处理，防止重复处理

## 使用建议

1. 在创建异步任务时配置 `callback_url`
2. 在回调处理程序中验证请求的合法性
3. 及时处理并保存任务结果
4. 实现日志记录便于问题排查
5. 设置合理的超时时间，避免回调处理时间过长
