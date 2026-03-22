# 构建前端
FROM node:22-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# 构建后端
FROM golang:1.24-alpine AS backend-builder
WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/main.go

# 最终镜像
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app

# 复制后端二进制文件
COPY --from=backend-builder /app/server .

# 复制前端构建产物
COPY --from=frontend-builder /app/frontend/dist ./static

EXPOSE 8080

CMD ["./server"]
