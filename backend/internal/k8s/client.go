package k8s

import (
	"fmt"
	"log"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"

	"github.com/yangqihuang/k8s-ui/internal/config"
)

// NewClient 根据 config 创建 Kubernetes 客户端
func NewClient(cfg *config.Config) (*kubernetes.Clientset, *rest.Config, error) {
	restConfig, err := getRESTConfig(cfg)
	if err != nil {
		return nil, nil, err
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("创建 k8s client 失败: %w", err)
	}

	log.Println("Kubernetes 客户端初始化成功")
	return clientset, restConfig, nil
}

// getRESTConfig 获取 Kubernetes REST 配置
func getRESTConfig(cfg *config.Config) (*rest.Config, error) {
	if cfg.InCluster {
		restConfig, err := rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("获取 in-cluster config 失败: %w", err)
		}
		log.Println("Kubernetes 客户端初始化成功 (in-cluster)")
		return restConfig, nil
	}

	kubeconfig := cfg.KubeConfig
	if kubeconfig == "" {
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		}
	}

	if kubeconfig == "" {
		return nil, fmt.Errorf("未指定 kubeconfig 路径且无法找到默认 kubeconfig")
	}

	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("构建 k8s config 失败: %w", err)
	}
	return restConfig, nil
}
