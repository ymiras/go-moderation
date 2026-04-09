# API 文档

## 目录

- [健康检查](#健康检查)
- [标准审核接口](#标准审核接口)
- [Dify 审核接口](#dify-审核接口)
- [认证](#认证)
- [错误响应](#错误响应)

---

## 健康检查

验证服务是否正常运行。

### 请求

```
GET /health
```

无需认证。

### 响应

```
HTTP/1.1 200 OK
X-Request-ID: 550e8400-e29b-41d4-a716-446655440000
```

**响应码：**

| 状态码 | 说明 |
|--------|------|
| 200 | 服务正常运行 |
| 500 | 服务内部错误 |

---

## 标准审核接口

适用于内部业务系统、网关集成等场景。

### 请求

```
POST /api/v1/text/moderation
Authorization: Bearer <api-key>
Content-Type: application/json
```

### 请求体

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| text | string | 是 | 待审核的文本内容 |
| point | string | 否 | 审核点，值为 `input`（默认）或 `output` |
| app_id | string | 否 | 应用标识符，用于追踪和日志 |

**示例：**

```json
{
  "text": "这是一段需要审核的文本内容",
  "point": "input",
  "app_id": "my-chatbot-app"
}
```

### 响应

| 字段 | 类型 | 说明 |
|------|------|------|
| flagged | boolean | 是否命中敏感内容 |
| action | string | 建议的动作：`pass`、`block` 或 `mask` |
| hits | array | 命中的敏感词列表 |
| latency_ms | float | 处理耗时（毫秒） |

**命中记录结构：**

| 字段 | 类型 | 说明 |
|------|------|------|
| word | string | 命中的敏感词 |
| type | string | 敏感词类型 |
| severity | string | 严重程度：`low`、`medium`、`high` |
| index | int | 匹配位置（起始索引） |
| length | int | 匹配长度 |

**示例响应：**

```json
{
  "flagged": true,
  "action": "block",
  "hits": [
    {
      "word": "敏感词",
      "type": "profanity",
      "severity": "high",
      "index": 5,
      "length": 3
    }
  ],
  "latency_ms": 1.45
}
```

---

## Dify 审核接口

严格遵循 Dify API Extension 规范。

### 请求

```
POST /dify/moderation
Authorization: Bearer <api-key>
Content-Type: application/json
```

### 请求体

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| point | string | 是 | 审核点类型 |
| params | object | 是 | Dify 扩展参数 |

**支持的 point 值：**

| 值 | 说明 |
|----|------|
| `app.moderation.input` | 审核用户输入 |
| `app.moderation.output` | 审核模型输出 |

**params 字段：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| app_id | string | 否 | 应用 ID |
| query | string | 条件 | 用户输入（point=input 时必填） |
| text | string | 条件 | 模型输出（point=output 时必填） |
| inputs | object | 否 | Dify 输入变量 |

**示例请求 - 审核输入：**

```json
{
  "point": "app.moderation.input",
  "params": {
    "app_id": "dify-chatbot",
    "query": "用户发送的消息内容"
  }
}
```

**示例请求 - 审核输出：**

```json
{
  "point": "app.moderation.output",
  "params": {
    "app_id": "dify-chatbot",
    "text": "AI 生成的回复内容"
  }
}
```

### 响应

Dify 接口返回严格遵循官方规范的 JSON：

| 字段 | 类型 | 说明 |
|------|------|------|
| flagged | boolean | 是否命中敏感内容 |
| action | string | 动作：`direct_output`（放行）或 `overridden`（拦截） |
| preset_response | string | 预设响应文本（当 action=overridden 时） |

**响应示例：**

```json
{
  "flagged": false,
  "action": "direct_output",
  "preset_response": ""
}
```

**拦截示例：**

```json
{
  "flagged": true,
  "action": "overridden",
  "preset_response": "内容包含敏感信息，请修改后重试"
}
```

---

## 认证

所有审核接口都需要 Bearer Token 认证。

### 认证头

```
Authorization: Bearer <your-api-key>
```

### 配置 API 密钥

通过配置文件或环境变量配置：

```yaml
auth:
  api_keys:
    - "your-first-key"
    - "your-second-key"
```

或环境变量：

```bash
MODERATION_AUTH_API_KEYS=key1,key2
```

### 认证失败

缺少或无效的 Token 返回 401：

```
HTTP/1.1 401 Unauthorized
Content-Type: application/json

{
  "error": "invalid or missing authorization token"
}
```

---

## 错误响应

### 错误格式

```json
{
  "error": "错误描述信息"
}
```

### 错误码

| HTTP 状态码 | 说明 |
|-------------|------|
| 400 | 请求参数错误（缺少必填字段、格式错误等） |
| 401 | 认证失败 |
| 429 | 请求频率超出限制 |
| 500 | 服务器内部错误 |

### 示例

**缺少 text 字段（400）：**

```json
{
  "error": "text is required"
}
```

**无效的 point 值（400）：**

```json
{
  "error": "point must be 'input' or 'output'"
}
```

**超出限流（429）：**

```json
{
  "error": "rate limit exceeded"
}
```

---

## 速率限制

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| rate | 100 | 每秒允许的请求数 |
| capacity | 200 | 令牌桶容量 |

超出限制返回 `429 Too Many Requests`，响应头包含 `Retry-After` 指示等待秒数。

---

## 延迟性能

典型延迟（单核 10k+ QPS）：

| 场景 | P50 | P95 | P99 |
|------|-----|-----|-----|
| 短文本（<100字） | <1ms | <2ms | <5ms |
| 长文本（1k字） | <5ms | <10ms | <15ms |
