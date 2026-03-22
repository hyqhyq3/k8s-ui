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

# 部署到 Kubernetes
deploy: build
	kubectl apply -f deploy/manifests/

# 删除部署
clean:
	kubectl delete -f deploy/manifests/ --ignore-not-found=true

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
