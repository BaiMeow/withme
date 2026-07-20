# ---------- 前端构建（产物直接写入 Go 的 embed 目录） ----------
FROM node:22-alpine AS frontend
WORKDIR /build/frontend
COPY frontend/package.json frontend/pnpm-lock.yaml frontend/.npmrc ./
# packageManager 字段锁定 pnpm 版本，保证与本地/CI 一致
RUN corepack enable && corepack install && pnpm install --frozen-lockfile
COPY frontend/ ./
RUN pnpm build

# ---------- 后端构建（前端产物经 go:embed 打进二进制） ----------
FROM golang:1.26-alpine AS backend
WORKDIR /build/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
COPY --from=frontend /build/backend/internal/web/dist ./internal/web/dist
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /withme-server ./cmd/server

# ---------- 运行时 ----------
FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=backend /withme-server ./withme-server
# 默认配置（无密钥）；生产配置挂载到 /app/config/config.yaml 覆盖
COPY backend/config/config.example.yaml ./config/config.yaml
VOLUME ["/app/data"]
EXPOSE 8080
CMD ["./withme-server"]
