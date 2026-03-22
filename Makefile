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

# 测试
test-backend:
	cd backend && go test ./...

test-frontend:
	cd frontend && npm test

.PHONY: build push deploy clean dev-backend dev-frontend test-backend test-frontend
