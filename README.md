# WithMe · 相亲资料生成器

输入一个网名，通过 Gemini（Google Search grounding）检索互联网公开足迹，生成「圈内人版 / 圈外人版」两版相亲资料，可保存并分享。

## 目录结构

```
withme/
├── backend/            # Go 后端（Gin + Gemini）
│   ├── cmd/server/     # 入口
│   ├── internal/
│   │   ├── api/        # HTTP 接口
│   │   ├── config/     # 配置加载
│   │   ├── generator/  # Gemini 生成
│   │   ├── model/      # 数据模型
│   │   └── store/      # 持久化（sqlite / mysql 双驱动）
│   └── config/config.yaml
└── frontend/           # Vue 3 + Vite 前端
    └── src/
        ├── views/      # HomeView(生成+历史) / ProfileView(分享页)
        └── components/ # ProfileCard(资料卡片)
```

## 本地开发

```bash
# 后端（默认 :8080，sqlite 存到 backend/data/withme.db）
cd backend && go run ./cmd/server

# 前端（:5173，/api 自动代理到 8080）
cd frontend && pnpm dev
```

前端产物通过 `go:embed` 打进后端二进制：`pnpm build` 的输出目录就是
`backend/internal/web/dist`，先构建前端再构建后端即为单文件分发。

## 线上部署

### 手动部署

```bash
cd frontend && pnpm build          # 产物直接输出到 Go 的 embed 目录
cd backend && go build -o withme-server ./cmd/server
./withme-server                    # 单二进制托管全部流量
```

### Docker / GHCR

推送到 main 分支后，GitHub Actions 自动构建并推送镜像到
`ghcr.io/baimeow/withme`（linux/amd64；tags：`latest` / `sha-xxx` / `v*` 语义化版本）。

镜像内不含密钥，将生产配置挂载进容器：

```bash
docker run -d -p 8080:8080 \
  -v ./config.yaml:/app/config/config.yaml:ro \
  -v withme-data:/app/data \
  ghcr.io/baimeow/withme:latest
```

或在仓库根目录准备好 `config.yaml` 后 `docker compose up -d`。

### 配置与密钥

密钥**不入库**：`backend/config/config.yaml` 已 gitignore，仓库内仅有
`config.example.yaml`，复制后填入真实配置：

```yaml
gemini:
  api_key: "你的密钥"
  model: "gemini-3.5-flash"
```

线上使用 MySQL：

```yaml
database:
  driver: mysql
  dsn: "user:password@tcp(127.0.0.1:3306)/withme?charset=utf8mb4&parseTime=true"
```

表结构会自动创建，sqlite / mysql 无需任何代码改动。

## API

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | `/api/generate` | 生成资料 `{username, version}` → `{id, profile}` |
| GET | `/api/profiles` | 最近生成列表（摘要） |
| GET | `/api/profiles/:id` | 按分享 ID 取资料（浏览数 +1） |

分享链接格式：`/p/{id}`（前端路由，SPA 回退处理）。
