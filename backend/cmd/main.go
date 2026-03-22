package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yangqihuang/k8s-ui/internal/config"
	"github.com/yangqihuang/k8s-ui/internal/handler"
	k8sclient "github.com/yangqihuang/k8s-ui/internal/k8s"
	"github.com/yangqihuang/k8s-ui/internal/service"
)

func main() {
	cfg := config.Load()

	// 初始化 K8s 客户端
	k8sClient, err := k8sclient.NewClient(cfg)
	if err != nil {
		log.Fatalf("Kubernetes 客户端初始化失败: %v", err)
	}

	k8sSvc := service.NewK8sService(k8sClient)
	h := handler.NewHandler(k8sSvc)

	r := gin.Default()

	// 健康检查
	r.GET("/health", h.Health)

	// API 路由组
	api := r.Group("/api/v1")
	{
		api.GET("/ping", h.Ping)
		api.GET("/stats", h.GetClusterStats)
		api.GET("/resources/:resource/:name", h.GetResourceYAML)
		api.GET("/namespaces", h.ListNamespaces)
		api.GET("/pods", h.ListPods)
		api.GET("/deployments", h.ListDeployments)
		api.GET("/statefulsets", h.ListStatefulSets)
		api.GET("/daemonsets", h.ListDaemonSets)
		api.GET("/configmaps", h.ListConfigMaps)
		api.GET("/secrets", h.ListSecrets)
		api.GET("/pvs", h.ListPersistentVolumes)
		api.GET("/pvcs", h.ListPersistentVolumeClaims)
		api.GET("/storageclasses", h.ListStorageClasses)
	}

	// 静态文件（Vite 构建产物在 assets 目录下）
	r.Static("/assets", "./static/assets")

	// 根级静态文件
	r.StaticFile("/favicon.svg", "./static/favicon.svg")
	r.StaticFile("/icons.svg", "./static/icons.svg")

	// SPA fallback - 所有未匹配路由返回 index.html
	r.NoRoute(func(c *gin.Context) {
		if c.Request.Method == http.MethodGet {
			c.File("./static/index.html")
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	})

	log.Printf("Server starting on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
