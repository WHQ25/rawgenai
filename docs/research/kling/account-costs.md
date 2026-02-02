# 账号信息查询

## 查询账号下资源包列表及余量

| 网络协议 | 请求地址 | 请求方法 | 请求格式 | 响应格式 |
|---------|---------|---------|---------|---------|
| https | /account/costs | GET | application/json | application/json |

## 请求头

| 字段 | 值 | 描述 |
|------|-----|------|
| Content-Type | application/json | 数据交换格式 |
| Authorization | 鉴权信息 | 鉴权信息 |

## 请求路径参数

| 字段 | 类型 | 必填 | 默认值 | 描述 |
|------|------|------|--------|------|
| start_time | int | 是 | 无 | 查询的开始时间，Unix时间戳，单位ms |
| end_time | int | 是 | 无 | 查询的结束时间，Unix时间戳，单位ms |
| resource_pack_name | string | 否 | 无 | 资源包名称，用于精准指定查询某个资源包 |

## 响应体

```json
{
    "code": 0,
    "message": "string",
    "request_id": "string",
    "data": {
        "code": 0,
        "msg": "string",
        "resource_pack_subscribe_infos": [
            {
                "resource_pack_name": "string",
                "resource_pack_id": "string",
                "resource_pack_type": "string",  // decreasing_total / constant_period
                "total_quantity": 200.0,
                "remaining_quantity": 118.0,
                "purchase_time": 1726124664368,
                "effective_time": 1726124664368,
                "invalid_time": 1727366400000,
                "status": "string"  // toBeOnline / online / expired / runOut
            }
        ]
    }
}
```

## 字段说明

### 资源包类型（resource_pack_type）

| 类型 | 描述 |
|------|------|
| decreasing_total | 总量递减型 |
| constant_period | 周期恒定型 |

### 资源包状态（status）

| 状态 | 描述 |
|------|------|
| toBeOnline | 待生效 |
| online | 生效中 |
| expired | 已到期 |
| runOut | 已用完 |

## 注意事项

- 该接口**免费调用**，用于查询账号下的资源包列表和余量
- 请注意控制请求速率，**QPS ≤ 1**
- start_time 和 end_time 必须同时提供
- resource_pack_name 为可选参数，可用于精准查询特定资源包
