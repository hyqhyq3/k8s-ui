package config

import (
	"os"
)

// Config 应用配置
type Config struct {
	Port        string
	KubeConfig  string
	InCluster   bool
}

// Load 加载配置
func Load() *Config {
	return &Config{
		Port:       getEnv("PORT", "8080"),
		KubeConfig: getEnv("KUBECONFIG", ""),
		InCluster:  getEnv("IN_CLUSTER", "false") == "true",
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
