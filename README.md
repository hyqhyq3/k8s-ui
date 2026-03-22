# k8s-ui

基于 Golang + React 实现的 Kubernetes Web 管理界面

## 项目结构

```
.
├── backend/           # Golang 后端 API
│   ├── cmd/          # 应用入口
│   ├── internal/     # 内部包
│   │   ├── config/   # 配置管理
│   │   ├── handler/  # HTTP 处理器
│   │   ├── service/  # 业务逻辑
│   │   └── model/    # 数据模型
│   └── go.mod        # Go 模块
├── frontend/         # React + TypeScript 前端
│   ├── src/          # 源代码
│   ├── public/       # 静态资源
│   └── package.json  # Node 依赖
├── deploy/           # Kubernetes 部署配置
│   └── manifests/    # K8s YAML 文件
├── Dockerfile        # 多阶段构建
└── Makefile          # 构建脚本
```

## 技术栈

- **后端**: Golang + Gin 框架
- **前端**: React + TypeScript + Vite
- **容器化**: Docker
- **部署**: Kubernetes

## 快速开始

### 本地开发

```bash
# 运行后端
cd backend && go run cmd/main.go

# 运行前端（新终端）
cd frontend && npm run dev
```

### Docker 构建

```bash
# 构建镜像
make build

# 推送镜像
make push image=your-registry/k8s-ui:latest
```

### Kubernetes 部署

```bash
# 部署到 K8s
make deploy

# 删除部署
make clean
```

## API 接口

- `GET /health` - 健康检查
- `GET /api/v1/ping` - 测试接口

## 开发计划

- [x] 项目基础框架
- [x] Docker 多阶段构建
- [x] Kubernetes 部署配置
- [ ] K8s 资源查看功能
- [ ] 前端 UI 组件
- [ ] 认证授权
