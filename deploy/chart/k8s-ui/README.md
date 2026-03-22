# k8s-ui Helm Chart

用于部署 k8s-ui 的 Helm Chart

## 前置要求

- Kubernetes 1.19+
- Helm 3.2.0+

## 安装

```bash
# 添加仓库（如果有）
# helm repo add myrepo https://charts.example.com

# 安装 Chart
helm install k8s-ui ./deploy/chart/k8s-ui -n k8s-ui --create-namespace

# 安装并指定镜像
helm install k8s-ui ./deploy/chart/k8s-ui \
  --set image.repository=myregistry/k8s-ui \
  --set image.tag=v1.0.0 \
  -n k8s-ui --create-namespace
```

## 升级

```bash
helm upgrade k8s-ui ./deploy/chart/k8s-ui -n k8s-ui
```

## 卸载

```bash
helm uninstall k8s-ui -n k8s-ui
```

## 配置参数

| 参数 | 描述 | 默认值 |
|------|------|--------|
| `replicaCount` | 副本数量 | `1` |
| `image.repository` | 镜像仓库 | `k8s-ui` |
| `image.tag` | 镜像标签 | `""` (使用 Chart appVersion) |
| `image.pullPolicy` | 镜像拉取策略 | `IfNotPresent` |
| `service.type` | 服务类型 | `ClusterIP` |
| `service.port` | 服务端口 | `80` |
| `ingress.enabled` | 启用 Ingress | `false` |
| `autoscaling.enabled` | 启用 HPA | `false` |
| `resources` | 资源限制 | 见 values.yaml |

## 示例

### 启用 Ingress

```bash
helm install k8s-ui ./deploy/chart/k8s-ui \
  --set ingress.enabled=true \
  --set ingress.hosts[0].host=k8s-ui.example.com \
  --set ingress.className=nginx \
  -n k8s-ui --create-namespace
```

### 启用自动扩缩容

```bash
helm install k8s-ui ./deploy/chart/k8s-ui \
  --set autoscaling.enabled=true \
  --set autoscaling.minReplicas=2 \
  --set autoscaling.maxReplicas=5 \
  -n k8s-ui --create-namespace
```

### 使用自定义 values 文件

```bash
helm install k8s-ui ./deploy/chart/k8s-ui -f my-values.yaml -n k8s-ui
```
