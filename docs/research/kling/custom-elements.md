# 主体创建

## 创建主体

| 网络协议 | 请求地址 | 请求方法 | 请求格式 | 响应格式 |
|---------|---------|---------|---------|---------|
| https | /v1/general/custom-elements | POST | application/json | application/json |

## 请求头

| 字段 | 值 | 描述 |
|------|-----|------|
| Content-Type | application/json | 数据交换格式 |
| Authorization | 鉴权信息 | 鉴权信息 |

## 请求体

| 字段 | 类型 | 必填 | 默认值 | 描述 |
|------|------|------|--------|------|
| element_name | string | 是 | 无 | 主体名称，不超过20字符 |
| element_description | string | 是 | 无 | 主体描述，不超过100字符 |
| element_frontal_image | string | 是 | 无 | 主体正面参考图（Base64编码或URL），.jpg/.jpeg/.png，≤10MB，≥300px，宽高比1:2.5~2.5:1 |
| element_refer_list | array | 是 | 无 | 主体其他参考列表（1-3张），格式：`[{"image_url": "url1"}, {"image_url": "url2"}]` |
| tag_list | array | 可选 | 空 | 标签列表，格式：`[{"tag_id": "o_101"}]` |

### 标签类型

| tag_id | 名称 | 描述 |
|--------|------|------|
| o_101 | 热梗 | 网络热梗或流行文化 |
| o_102 | 人物 | 人物角色 |
| o_103 | 动物 | 动物或宠物 |
| o_104 | 道具 | 道具或物品 |
| o_105 | 服饰 | 服装或穿戴 |
| o_106 | 场景 | 背景场景 |
| o_107 | 特效 | 特殊效果 |
| o_108 | 其他 | 其他类别 |

## 响应体

```json
{
    "code": 0,
    "message": "string",
    "request_id": "string",
    "data": {
        "element_id": "long",
        "element_name": "string",
        "element_description": "string",
        "element_frontal_image": "image_url_0",
        "element_refer_list": [
            {"image_url": "image_url_1"},
            {"image_url": "image_url_2"},
            {"image_url": "image_url_3"}
        ],
        "tag_list": [
            {
                "id": "o_101",
                "name": "string",
                "description": "string"
            }
        ],
        "owned_by": "string"  // kling为官方主体库，其他为创作者ID
    }
}
```

---

## 查询自定义主体（列表）

| 网络协议 | 请求地址 | 请求方法 |
|---------|---------|---------|
| https | /v1/general/custom-elements | GET |

### 查询参数

```
/v1/general/custom-elements?pageNum=1&pageSize=30
```

| 字段 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| pageNum | int | 1 | 页码，[1,1000] |
| pageSize | int | 30 | 每页数据量，[1,500] |

### 响应体

```json
{
    "code": 0,
    "message": "string",
    "request_id": "string",
    "data": [
        {
            "element_id": "long",
            "element_name": "string",
            "element_description": "string",
            "element_frontal_image": "image_url_0",
            "element_refer_list": [
                {"image_url": "image_url_1"},
                {"image_url": "image_url_2"},
                {"image_url": "image_url_3"}
            ],
            "tag_list": [
                {
                    "id": "o_101",
                    "name": "string",
                    "description": "string"
                }
            ],
            "owned_by": "string"
        }
    ]
}
```

---

## 查询官方主体（列表）

| 网络协议 | 请求地址 | 请求方法 |
|---------|---------|---------|
| https | /v1/general/presets-elements | GET |

### 查询参数

```
/v1/general/presets-elements?pageNum=1&pageSize=30
```

| 字段 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| pageNum | int | 1 | 页码，[1,1000] |
| pageSize | int | 30 | 每页数据量，[1,500] |

### 响应体

```json
{
    "code": 0,
    "message": "string",
    "request_id": "string",
    "data": [
        {
            "element_id": "long",
            "element_name": "string",
            "element_description": "string",
            "element_frontal_image": "image_url_0",
            "element_refer_list": [
                {"image_url": "image_url_1"},
                {"image_url": "image_url_2"},
                {"image_url": "image_url_3"}
            ],
            "tag_list": [
                {
                    "id": "o_101",
                    "name": "string",
                    "description": "string"
                }
            ],
            "final_unit_deduction": "string",
            "owned_by": "kling"
        }
    ]
}
```

---

## 删除自定义主体

| 网络协议 | 请求地址 | 请求方法 | 请求格式 | 响应格式 |
|---------|---------|---------|---------|---------|
| https | /v1/general/delete-elements | POST | application/json | application/json |

### 请求头

| 字段 | 值 | 描述 |
|------|-----|------|
| Content-Type | application/json | 数据交换格式 |
| Authorization | 鉴权信息 | 鉴权信息 |

### 请求体

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| element_id | string | 是 | 要删除的主体ID，仅支持删除自定义主体 |

### 响应体

```json
{
    "code": 0,
    "message": "string",
    "request_id": "string",
    "data": {
        "task_id": "string",
        "task_status": "string"  // submitted/processing/succeed/failed
    }
}
```
