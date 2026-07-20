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

## 线上部署

### 手动部署

```bash
cd frontend && pnpm build          # 产物输出到 frontend/dist
cd backend && go build -o withme-server ./cmd/server
./withme-server                    # 后端直接托管 dist，单进程服务全部流量
```

### Docker / GHCR

推送到 main 分支后，GitHub Actions 自动构建并推送镜像到
`ghcr.io/baimeow/withme`（tags：`latest` / `sha-xxx` / `v*` 语义化版本）。

```bash
docker run -d -p 8080:8080 \
  -e GEMINI_API_KEY=你的密钥 \
  -v withme-data:/app/data \
  ghcr.io/baimeow/withme:latest
```

或使用 compose：`GEMINI_API_KEY=... docker compose up -d`

### 配置与密钥

密钥**不入库**：`backend/config/config.yaml` 已 gitignore，仓库内仅有
`config.example.yaml`。所有配置均可用环境变量覆盖：

| 环境变量 | 说明 |
|---|---|
| `GEMINI_API_KEY` | Gemini 密钥（必需） |
| `GEMINI_MODEL` | 模型 |
| `DATABASE_DRIVER` | `sqlite`（默认）/ `mysql` |
| `DATABASE_DSN` | 数据库连接串 |
| `SERVER_PORT` | 端口，默认 8080 |
| `FRONTEND_DIST` | 前端产物目录 |
| `CONFIG_PATH` | 配置文件路径，默认 ./config/config.yaml |

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
