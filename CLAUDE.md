# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Kubernetes Web 管理界面，采用前后端分离架构。后端通过 Gin 框架与 Kubernetes API 交互，前端使用 React + Ant Design + Vite 构建。支持 K8s 资源查看（Pod、Deployment、Service、Ingress、Node 等）和 Helm Release 管理。

## Build & Development Commands

```bash
# 后端开发（默认 :8080）
cd backend && go run cmd/main.go

# 前端开发（Vite dev server，代理 /api 到 :8080）
cd frontend && npm run dev

# 前端构建（tsc 检查 + vite build）
cd frontend && npm run build

# 前端 lint
cd frontend && npm run lint

# 后端测试
cd backend && go test ./...

# 后端单测（带覆盖率和详细输出）
cd backend && go test ./... -v -coverprofile=coverage.out

# 前端 E2E 测试
cd frontend && npm run test:e2e

# Docker 构建
make build

# Docker 推送
make push image=<registry>/<name>:<tag>

# Helm 部署/卸载/模板渲染
make helm-install
make helm-uninstall
make helm-template
```

## Architecture

### Backend (`backend/`)

Go + Gin，模块路径 `github.com/yangqihuang/k8s-ui`，Go 1.25+。

```
cmd/main.go                # 入口：初始化 K8s client、Helm driver、注册所有路由
internal/
├── config/config.go       # 环境变量：PORT, KUBECONFIG, IN_CLUSTER
├── k8s/client.go          # K8s clientset 初始化（支持 kubeconfig 和 in-cluster）
├── helm/
│   ├── driver.go          # Helm SDK driver 封装
│   └── options.go         # Helm 配置选项
├── handler/
│   ├── handler.go         # K8s 资源 HTTP 处理器
│   └── helm_handler.go    # Helm Release/Repo HTTP 处理器
├── service/
│   ├── k8s_service.go     # K8s 业务逻辑（资源 CRUD、YAML 导出、日志、事件）
│   └── helm_service.go    # Helm 业务逻辑
└── model/                 # 数据模型（当前为空，模型定义在 service 层）
```

- **API 路由前缀**: `/api/v1/`，K8s 资源路由和 Helm 路由分为两个 group
- **SPA 静态文件**: 后端同时 serve 前端构建产物，`/assets` 静态目录 + `NoRoute` fallback 到 `index.html`
- **数据模型**: 所有 K8s 资源的 Info/Detail 结构体直接定义在 `service/k8s_service.go` 中，不在独立 model 包里
- **依赖**: `k8s.io/client-go`, `helm.sh/helm/v3`, `gin-gonic/gin`, `sigs.k8s.io/yaml`

### Frontend (`frontend/`)

React 19 + TypeScript + Ant Design 6 + Vite 8。

```
src/
├── api/
│   ├── client.ts           # HTTP 封装：get/post/del，统一 unwrap ApiResponse<T>
│   ├── k8s.ts              # K8s 资源 API 调用
│   └── helm.ts             # Helm API 调用
├── hooks/
│   └── useResourceList.ts  # 通用资源列表 Hook：加载、namespace 过滤、客户端搜索
├── components/             # 复用组件：YAMLViewer, InstallModal, UpgradeModal, ValuesViewer
├── layouts/
│   └── MainLayout.tsx      # Ant Design ProLayout 侧边栏布局
├── pages/                  # 页面组件，每个 K8s 资源类型一个页面
└── types/                  # TypeScript 类型定义（k8s.ts, helm.ts）
```

- **路由**: `react-router-dom` v7，所有页面嵌套在 `MainLayout` 下
- **API 代理**: Vite dev server 将 `/api` 请求代理到 `http://localhost:8080`
- **国际化**: Ant Design 使用 `zh_CN` locale
- **列表页模式**: 大多数资源列表页使用 `useResourceList` Hook + Ant Design Table，支持 namespace 筛选和关键词搜索

### Deploy

```
deploy/
├── manifests/              # 原生 K8s YAML（namespace, rbac, deployment, service）
└── chart/k8s-ui/           # Helm Chart
```

- **Docker 多阶段构建**: `frontend-builder`(node:22-alpine) → `backend-builder`(golang:1.25-alpine) → `alpine:latest`
- 前端构建产物输出到 `/app/frontend/dist`，后端复制到 `./static` 目录
- 后端编译为 CGO_ENABLED=0 的静态二进制

## Key Conventions

- **后端端口**: 默认 8080，通过 `PORT` 环境变量配置
- **K8s 集群访问**: 支持 kubeconfig 文件（`KUBECONFIG`）和 in-cluster 模式（`IN_CLUSTER=true`），默认使用 `~/.kube/config`
- **API 响应格式**: 统一 `{ "data": ... }` 或 `{ "error": "..." }` JSON 格式
- **前端 UI 语言**: 中文
