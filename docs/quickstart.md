# 快速开始指南

本文档帮助你快速部署和运行 Dify Moderation 服务。

## 前置要求

- **Go 1.23+** - [安装指南](https://go.dev/doc/install)
- **Docker** (可选) - [安装指南](https://docs.docker.com/get-docker/)

## 安装步骤

### 1. 克隆项目

```bash
git clone https://github.com/ymiras/go-moderation.git
cd go-moderation
```

### 2. 安装依赖

```bash
go mod download
```

### 3. 配置服务

复制环境变量模板并修改：

```bash
cp .env.example .env
```

编辑 `.env` 文件，添加你的 API 密钥：

```bash
MODERATION_AUTH_API_KEYS=your-secret-key-here
```

### 4. 启动服务

**方式一：直接运行**

```bash
go run cmd/server/main.go
```

**方式二：使用 Makefile**

```bash
make run
```

**方式三：使用 Docker**

```bash
# 构建镜像
docker build -t go-moderation:latest .

# 运行容器
docker run -d \
  --name go-moderation \
  -p 8080:8080 \
  -v $(pwd)/configs:/app/configs:ro \
  -e MODERATION_AUTH_API_KEYS=your-secret-key \
  go-moderation:latest
```

**方式四：使用 Docker Compose**

```bash
docker-compose up -d
```

### 5. 验证服务

检查服务是否正常运行：

```bash
curl http://localhost:8080/health
```

正常返回 200 OK。

## 测试审核功能

### 使用标准接口

```bash
curl -X POST http://localhost:8080/api/v1/text/moderation \
  -H "Authorization: Bearer your-secret-key" \
  -H "Content-Type: application/json" \
  -d '{"text": "这是一段正常的文本内容", "point": "input"}'
```

正常文本响应：

```json
{"flagged":false,"action":"pass","hits":[],"latency_ms":0.85}
```

### 测试 Dify 接口

```bash
curl -X POST http://localhost:8080/dify/moderation \
  -H "Authorization: Bearer your-secret-key" \
  -H "Content-Type: application/json" \
  -d '{
    "point": "app.moderation.input",
    "params": {
      "query": "用户输入的测试内容"
    }
  }'
```

## 配置敏感词库

### 编辑词库文件

编辑 `configs/wordlist/default.csv`：

```csv
# 格式: 词语,类型,严重程度,动作
敏感词,profanity,high,block
广告内容,spam,medium,review
```

**字段说明：**

| 字段 | 说明 | 可选值 |
|------|------|--------|
| 词语 | 敏感词 | - |
| 类型 | 分类标识 | profanity, spam, hate, violence, political |
| 严重程度 | 严重等级 | low, medium, high |
| 动作 | 建议处理 | block, review, pass |

### 使用自定义词库路径

```bash
MODERATION_WORD_BANK_PATH=/path/to/your/wordlist.csv go run cmd/server/main.go
```

## 配置正则规则

编辑 `configs/regex/custom_rules.yaml`：

```yaml
patterns:
  - pattern: '[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}'
    name: email_pattern
    type: privacy
    severity: medium
    action: review
```

## 常见问题

### Q: 服务启动失败，提示端口被占用

**A:** 修改监听端口：

```bash
MODERATION_SERVER_PORT=8081 go run cmd/server/main.go
```

### Q: 认证失败，返回 401

**A:** 检查 API 密钥配置：

1. 确认 `.env` 文件中设置了 `MODERATION_AUTH_API_KEYS`
2. 请求时使用正确的 Bearer Token
3. 注意多个密钥用逗号分隔

### Q: 如何开启调试日志？

**A:** 设置日志级别：

```bash
MODERATION_LOG_LEVEL=debug go run cmd/server/main.go
```

## 停止服务

**直接运行：**

```
Ctrl+C
```

**Docker：**

```bash
docker stop go-moderation
docker rm go-moderation
```

**Docker Compose：**

```bash
docker-compose down
```

## 下一步

- 阅读 [API 文档](api.md) 了解更多接口详情
- 查看 [完整配置说明](../configs/default.yaml)
- 参考 [Dify 集成指南](../README.md#dify-集成指南) 接入 Dify 平台
