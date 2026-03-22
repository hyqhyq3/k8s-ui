# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Kubernetes Web 管理界面，采用前后端分离架构。后端通过 Gin 框架与 Kubernetes API 交互，前端使用 React + Vite 构建。

## Build & Development Commands

```bash
# 后端开发（默认 :8080）
cd backend && go run cmd/main.go

# 前端开发（Vite dev server）
cd frontend && npm run dev

# 前端构建（tsc 检查 + vite build）
cd frontend && npm run build

# 前端 lint
cd frontend && npm run lint

# 后端测试
cd backend && go test ./...

# Docker 构建
make build

# Docker 推送
make push image=<registry>/<name>:<tag>

# K8s 部署
make deploy

# K8s 清理
make clean
```

## Architecture

```
backend/
├── cmd/main.go              # 入口，Gin 路由注册
├── internal/
│   ├── config/config.go     # 环境变量配置（PORT, KUBECONFIG, IN_CLUSTER）
│   ├── handler/handler.go   # HTTP 处理器（当前仅 Health/Ping）
│   ├── service/             # 业务逻辑（待实现，K8s client 交互层）
│   └── model/               # 数据模型（待实现）
└── pkg/                     # 公共工具包

frontend/
├── src/                     # React 19 + TypeScript 源码
├── vite.config.ts           # Vite 配置
└── package.json             # React 19, Vite 8, ESLint 9

deploy/
├── manifests/               # 原生 K8s YAML（namespace, rbac, deployment, service）
└── chart/k8s-ui/            # Helm Chart
```

## Key Conventions

- **Go module**: `github.com/yangqihuang/k8s-ui`，Go 1.25+
- **API 路由前缀**: `/api/v1/`
- **后端端口**: 默认 8080，通过 `PORT` 环境变量配置
- **K8s 集群访问**: 支持 kubeconfig 文件（`KUBECONFIG`）和 in-cluster 模式（`IN_CLUSTER=true`）
- **Docker 多阶段构建**: 前端构建产物输出到 `/app/static`，后端编译为静态二进制
- **前端暂未集成 Ant Design**，当前使用 Vite 默认模板样式
