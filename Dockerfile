# ---------- 前端构建 ----------
FROM node:22-alpine AS frontend
WORKDIR /build
COPY frontend/package.json frontend/pnpm-lock.yaml frontend/.npmrc ./
# packageManager 字段锁定 pnpm 版本，保证与本地/CI 一致
RUN corepack enable && corepack install && pnpm install --frozen-lockfile
COPY frontend/ ./
RUN pnpm build

# ---------- 后端构建 ----------
FROM golang:1.26-alpine AS backend
WORKDIR /build
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /withme-server ./cmd/server

# ---------- 运行时 ----------
FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=backend /withme-server ./withme-server
COPY --from=frontend /build/dist ./frontend/dist
COPY backend/config/config.example.yaml ./config/config.yaml

# 配置均可用环境变量覆盖：GEMINI_API_KEY / DATABASE_DRIVER / DATABASE_DSN / SERVER_PORT 等
ENV FRONTEND_DIST=./frontend/dist \
    DATABASE_DSN=./data/withme.db
VOLUME ["/app/data"]
EXPOSE 8080
CMD ["./withme-server"]
