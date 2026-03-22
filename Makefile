# 镜像配置
IMAGE_REGISTRY ?= docker.io
IMAGE_NAME ?= k8s-ui
IMAGE_TAG ?= latest
IMAGE ?= $(IMAGE_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)

# 构建 Docker 镜像
build:
	docker build -t $(IMAGE) .

# 推送 Docker 镜像
push:
	docker push $(IMAGE)

# 部署到 Kubernetes (原生 YAML)
deploy: build
	kubectl apply -f deploy/manifests/

# 删除部署 (原生 YAML)
clean:
	kubectl delete -f deploy/manifests/ --ignore-not-found=true

# Helm 部署
helm-install:
	helm upgrade --install k8s-ui ./deploy/chart/k8s-ui \
		--namespace k8s-ui --create-namespace \
		--set image.repository=$(IMAGE_REGISTRY)/$(IMAGE_NAME) \
		--set image.tag=$(IMAGE_TAG)

# Helm 卸载
helm-uninstall:
	helm uninstall k8s-ui --namespace k8s-ui --ignore-not-found

# Helm 模板渲染（调试）
helm-template:
	helm template k8s-ui ./deploy/chart/k8s-ui \
		--namespace k8s-ui \
		--set image.repository=$(IMAGE_REGISTRY)/$(IMAGE_NAME) \
		--set image.tag=$(IMAGE_TAG)

# 本地开发 - 运行后端
dev-backend:
	cd backend && go run cmd/main.go

# 本地开发 - 运行前端
dev-frontend:
	cd frontend && npm run dev

# 测试 - 后端 API 测试
test-backend:
	cd backend && go test ./... -v

# 测试 - 后端覆盖率测试
test-backend-coverage:
	cd backend && go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out -o coverage.html

# 测试 - 前端 E2E 测试
test-frontend-e2e:
	cd frontend && npm run test:e2e

# 测试 - 前端 E2E 测试（带 UI）
test-frontend-e2e-ui:
	cd frontend && npm run test:e2e:ui

# 测试 - 前端 E2E 测试报告
test-frontend-e2e-report:
	cd frontend && npm run test:e2e:report

# 测试 - 运行所有测试
test-all: test-backend test-frontend-e2e

# 安装测试依赖
install-test-deps:
	cd frontend && npm install
	cd frontend && npx playwright install chromium

.PHONY: build push deploy clean dev-backend dev-frontend test-backend test-backend-coverage test-frontend-e2e test-frontend-e2e-ui test-frontend-e2e-report test-all install-test-deps
