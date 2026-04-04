# Go Moderation

高性能、双路由架构的内容安全审核服务。原生兼容 Dify API Extension，同时提供标准 REST API，适用于 AI 应用、LLM 网关及企业级文本过滤场景。

**Go 1.23+** | **MIT License** | **Dify API Extension Compatible** | **Docker Ready**

---

## 核心特性

- 双路由架构：标准接口 `/api/moderate` + Dify 专属适配 `/dify/moderation`，核心引擎完全解耦
- 极致性能：默认 AC 自动机匹配，单核 10k+ QPS，P99 延迟 < 15ms
- 词库热更新：原子指针替换 + 双缓冲切换，零停机、零误判
- 策略可插拔：支持 Chain / Parallel / Weighted 执行模式，混合 AC / Regex / 外部 API
- 生产可观测：结构化日志 + 健康检查
- 开箱即用：Docker / Docker Compose / Makefile 部署模板，内置优雅降级与限流

---

## 项目结构

```
go-moderation/
├── cmd/
│   └── server/main.go              # 服务入口
├── internal/
│   ├── adapter/
│   │   ├── dify/                  # Dify 协议适配层
│   │   └── standard/              # 标准 REST 适配层
│   ├── config/                     # 配置管理
│   ├── engine/                     # 核心审核引擎
│   ├── matcher/                    # 匹配器插件 (AC/Regex/External)
│   ├── middleware/                 # 中间件 (Auth/RateLimit/Logger)
│   ├── model/                      # 领域模型
│   └── storage/                    # 存储层 (WordBank/Cache)
├── pkg/
│   ├── logger/                     # Zap 日志封装
│   ├── metrics/                    # Prometheus 指标
│   └── util/                       # 工具函数
├── configs/
│   ├── default.yaml                # 基线配置
│   ├── wordlist/default.csv        # 敏感词库
│   └── regex/custom_rules.yaml      # 正则规则
├── scripts/                        # 运维脚本
├── Dockerfile                      # 多阶段构建
├── docker-compose.yml             # Docker 编排
└── Makefile                       # 构建命令
```

---

## 快速开始

### 前置要求

- Go 1.23+
- Docker (可选)

### 本地运行

```bash
# 克隆项目
git clone https://github.com/ymiras/go-moderation.git
cd go-moderation

# 下载依赖
go mod download

# 运行服务
go run cmd/server/main.go

# 或使用 Makefile
make run
```

服务启动于 `http://localhost:8080`

### Docker 运行

```bash
# 构建镜像
docker build -t dify-moderation:latest .

# 运行容器
docker run -d \
  --name dify-moderation \
  -p 8080:8080 \
  -v $(pwd)/configs:/app/configs:ro \
  dify-moderation:latest

# 或使用 docker-compose
docker-compose up -d
```

---

## API 参考

### 健康检查

```
GET /health
```

无需鉴权，返回 200 OK 表示服务正常运行。

**响应：**
```json
HTTP/1.1 200 OK
X-Request-ID: <uuid>
```

---

### 标准审核接口

```
POST /api/moderate
Authorization: Bearer <api-key>
Content-Type: application/json
```

**请求体：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| text | string | 是 | 待审核文本 |
| point | string | 否 | 审核点：`input` 或 `output`，默认 `input` |
| app_id | string | 否 | 应用 ID |

```json
{
  "text": "待审核文本内容",
  "point": "input",
  "app_id": "my-app"
}
```

**响应：**

| 字段 | 类型 | 说明 |
|------|------|------|
| flagged | bool | 是否命中敏感词 |
| action | string | 动作：`pass` / `block` / `mask` |
| hits | array | 命中记录列表 |
| latency_ms | float | 处理延迟（毫秒） |

```json
{
  "flagged": false,
  "action": "pass",
  "hits": [],
  "latency_ms": 1.23
}
```

---

### Dify 审核接口

```
POST /dify/moderation
Authorization: Bearer <api-key>
Content-Type: application/json
```

**请求体：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| point | string | 是 | 审核点：`app.moderation.input` 或 `app.moderation.output` |
| params | object | 是 | Dify 扩展参数 |
| params.query | string | 否 | 输入内容（point=input 时） |
| params.text | string | 否 | 输出内容（point=output 时） |
| params.app_id | string | 否 | 应用 ID |

```json
{
  "point": "app.moderation.input",
  "params": {
    "app_id": "my-app",
    "query": "用户输入内容"
  }
}
```

**响应：**

```json
{
  "flagged": false,
  "action": "direct_output",
  "preset_response": ""
}
```

---

## 配置说明

### 配置文件

默认配置文件位于 `configs/default.yaml`。

### 环境变量

所有配置都支持 `MODERATION_` 前缀的环境变量覆盖：

| 变量 | 默认值 | 说明 |
|------|--------|------|
| MODERATION_SERVER_HOST | 0.0.0.0 | 服务监听地址 |
| MODERATION_SERVER_PORT | 8080 | 服务监听端口 |
| MODERATION_SERVER_READ_TIMEOUT | 10s | 读超时 |
| MODERATION_SERVER_WRITE_TIMEOUT | 10s | 写超时 |
| MODERATION_LOG_LEVEL | info | 日志级别 |
| MODERATION_LOG_ENCODING | json | 日志编码 |
| MODERATION_AUTH_API_KEYS | [] | API 密钥列表 |
| MODERATION_RATELIMIT_RATE | 100 | 限流速率（请求/秒） |
| MODERATION_RATELIMIT_CAPACITY | 200 | 限流容量 |
| MODERATION_MODERATION_PIPELINE_MODE | chain | 执行模式 |
| MODERATION_MATCHERS_AC_ENABLED | true | 启用 AC 自动机 |
| MODERATION_MATCHERS_REGEX_ENABLED | false | 启用正则匹配 |
| MODERATION_MATCHERS_EXTERNAL_ENABLED | false | 启用外部 API |

### 配置示例

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: "10s"
  write_timeout: "10s"

auth:
  api_keys:
    - "your-secret-key-1"
    - "your-secret-key-2"

ratelimit:
  rate: 100
  capacity: 200

moderation:
  pipeline_mode: "chain"
  weighted_threshold: 0.5
  fallback_action: "pass"
```

---

## Dify 集成指南

1. 启动服务后，配置服务地址
2. 进入 Dify 后台：工作空间 -> 应用 -> API 扩展 -> 新建扩展
3. 填写配置：
   - 服务器 URL: `http://your-domain/dify/moderation`
   - 鉴权方式: Bearer Token
   - Token: 与 `MODERATION_AUTH_API_KEYS` 中的某个值一致
4. 启用审查：在 Chatflow / 助手 / Agent 中开启内容审查

---

## 构建和测试

```bash
# 构建二进制
make build

# 运行测试
make test

# 清理
make clean

# Docker 构建
make docker-build
make docker-run
```

---

## 运维脚本

| 脚本 | 说明 |
|------|------|
| scripts/health-check.sh | 健康检查 |
| scripts/benchmark.sh | 性能测试 |
| scripts/deploy.sh | Docker 部署 |

---

## 许可证

MIT License
