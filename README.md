# Dify Moderation Plugin (Go)

高性能、双路由架构的内容安全审核服务。原生兼容 Dify API Extension，同时提供标准 REST API，适用于 AI 应用、LLM 网关及企业级文本过滤场景。

状态徽章：Go 1.21+ | MIT License | Dify API Extension Compatible | Docker Ready

---

## 核心特性

- 双路由架构：标准接口 /api/v1/moderate + Dify 专属适配 /dify/，核心引擎完全解耦
- 极致性能：默认 AC 自动机匹配，单核 10k+ QPS，P99 延迟 < 15ms
- 词库热更新：原子指针替换 + 双缓冲切换，零停机、零误判
- 策略可插拔：支持 Chain / Parallel / Weighted 执行模式，混合 Bloom / Regex / 外部 API
- 生产可观测：Prometheus 指标 + Zap 结构化日志 + K8s Health Probes
- 开箱即用：Docker / Docker Compose / K8s 部署模板，内置优雅降级与限流

---

## 架构设计

请求流向：

1. Dify 平台 -> POST /dify/ -> Dify Adapter
2. 其他系统 -> POST /api/v1/moderate -> Standard Handler
3. 两者 -> 协议转换层 -> Core Moderation Engine
4. 核心引擎 -> 策略编排 Pipeline -> 匹配器集群 (AC 自动机/布隆过滤器/正则引擎/外部审核 API)
5. 输出 -> 异步审计日志 + Prometheus 指标

双路由说明：

| 路由 | 路径 | 协议 | 适用场景 | 鉴权变量 |
|------|------|------|----------|----------|
| Dify 插件 | POST /dify/ | 严格遵循 Dify app.moderation.input/output 规范 | Dify 工作流/助手内容审查 | MODERATION_DIFY_TOKEN |
| 标准接口 | POST /api/v1/moderate | 扁平化 RESTful {text, point, app_id} | 内部业务/其他 AI 平台/网关集成 | MODERATION_API_TOKEN |

---

## 快速开始

使用 Docker 运行：

```bash
docker run -d \
  --name dify-moderation \
  -p 8080:8080 \
  -e MODERATION_DIFY_TOKEN="your-dify-secret" \
  -e MODERATION_API_TOKEN="your-api-secret" \
  -e WORD_BANK_PATH="/app/configs/wordlist/default.csv" \
  ghcr.io/your-org/dify-moderation:latest
```

本地开发：

```bash
git clone https://github.com/ymiras/dify-moderation.git
cd dify-moderation
go run cmd/server/main.go
```

服务启动于 http://localhost:8080

---

## Dify 集成指南

1. 复制服务地址：http://yourdomain/dify
2. 进入 Dify 后台：工作空间 -> API 扩展 -> 新建扩展
3. 填写配置：
   - 服务器 URL: http://yourdomain/dify
   - 鉴权方式: Bearer Token
   - Token: 与启动变量 MODERATION_DIFY_TOKEN 一致
4. 启用审查：在 Chatflow / 助手 / Agent 中开启：
   - 内容审查 -> 审查输入内容 (app.moderation.input)
   - 内容审查 -> 审查输出内容 (app.moderation.output)

提示：插件已原生支持 Dify 流式输出分段审查（每 100 字符触发一次），无需额外配置。

---

## API 参考

标准接口：

请求：

```
POST /api/v1/moderate
Authorization: Bearer {MODERATION_API_TOKEN}
Content-Type: application/json
```

请求体：

```json
{
  "text": "待审核文本",
  "point": "input",
  "app_id": "optional"
}
```

响应：

```json
{
  "flagged": true,
  "action": "block",
  "payload": { "masked_text": "替换后的内容" },
  "hits": [{ "word": "敏感词", "type": "keyword", "severity": "high" }],
  "latency_ms": 4.2
}
```

Dify 接口：

请求：

```
POST /dify/
Authorization: Bearer {MODERATION_DIFY_TOKEN}
Content-Type: application/json
```

请求体：

```json
{
  "point": "app.moderation.input",
  "params": {
    "app_id": "app-xxx",
    "query": "用户输入内容"
  }
}
```

响应：严格匹配 Dify 官方规范（direct_output / overridden）。

---

## 配置说明

支持环境变量覆盖或挂载 configs/default.yaml：

| 变量 | 默认值 | 说明 |
|------|--------|------|
| SERVER_PORT | 8080 | 服务监听端口 |
| MODERATION_DIFY_TOKEN | "" | Dify 路由独立 Token |
| MODERATION_API_TOKEN | "" | 标准 API 独立 Token |
| PIPELINE_MODE | chain | 执行模式：chain \| parallel \| weighted |
| MATCHER_AC_ENABLED | true | 启用 AC 自动机 |
| WORD_BANK_SOURCE | file | 词库源：file \| redis \| mysql |
| FALLBACK_ACTION | pass | 服务异常时默认动作：pass \| block |

---

## 项目结构

```
dify-moderation/
├── cmd/server/main.go          # 启动入口
├── internal/
│   ├── engine/service.go       # 平台无关核心引擎
│   ├── adapter/
│   │   ├── dify/               # Dify 协议适配层
│   │   └── standard/           # 标准 REST 适配层
│   ├── matcher/                # 算法插件 (AC/Regex/Bloom)
│   ├── storage/wordbank.go     # 词库热更新管理
│   └── middleware/             # 鉴权/限流/日志
├── configs/default.yaml        # 配置模板
├── Dockerfile & docker-compose.yml
└── README.md
```

---

## 生产部署

Docker Compose 示例：

```yaml
version: '3.8'
services:
  moderation:
    image: ghcr.io/ymiras/dify-moderation:latest
    ports: ["8080:8080"]
    env_file: .env
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:8080/health"]
      interval: 10s
      timeout: 3s
      retries: 3
```

Kubernetes：提供完整 deployment.yaml、service.yaml、configmap.yaml，支持 HPA 自动扩缩容与 Prometheus 指标采集。详见 k8s/ 目录。
