package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yangqihuang/k8s-ui/internal/service"
)

// Handler HTTP 处理器
type Handler struct {
	k8sService *service.K8sService
}

// NewHandler 创建处理器
func NewHandler(k8sService *service.K8sService) *Handler {
	return &Handler{k8sService: k8sService}
}

// Health 健康检查
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}

// Ping 测试接口
func (h *Handler) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

// ListNamespaces 获取 namespace 列表
func (h *Handler) ListNamespaces(c *gin.Context) {
	namespaces, err := h.k8sService.ListNamespaces(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": namespaces})
}

// ListPods 获取 pod 列表
func (h *Handler) ListPods(c *gin.Context) {
	namespace := c.Query("namespace")
	pods, err := h.k8sService.ListPods(c.Request.Context(), namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": pods})
}
