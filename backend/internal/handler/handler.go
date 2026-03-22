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

// ListDeployments 获取 deployment 列表
func (h *Handler) ListDeployments(c *gin.Context) {
	namespace := c.Query("namespace")
	deployments, err := h.k8sService.ListDeployments(c.Request.Context(), namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": deployments})
}

// ListStatefulSets 获取 statefulset 列表
func (h *Handler) ListStatefulSets(c *gin.Context) {
	namespace := c.Query("namespace")
	statefulsets, err := h.k8sService.ListStatefulSets(c.Request.Context(), namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": statefulsets})
}

// ListConfigMaps 获取 configmap 列表
func (h *Handler) ListConfigMaps(c *gin.Context) {
	namespace := c.Query("namespace")
	configmaps, err := h.k8sService.ListConfigMaps(c.Request.Context(), namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": configmaps})
}

// ListSecrets 获取 secret 列表
func (h *Handler) ListSecrets(c *gin.Context) {
	namespace := c.Query("namespace")
	secrets, err := h.k8sService.ListSecrets(c.Request.Context(), namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": secrets})
}

// ListPersistentVolumes 获取 persistentvolume 列表
func (h *Handler) ListPersistentVolumes(c *gin.Context) {
	pvs, err := h.k8sService.ListPersistentVolumes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": pvs})
}

// ListPersistentVolumeClaims 获取 persistentvolumeclaim 列表
func (h *Handler) ListPersistentVolumeClaims(c *gin.Context) {
	namespace := c.Query("namespace")
	pvcs, err := h.k8sService.ListPersistentVolumeClaims(c.Request.Context(), namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": pvcs})
}

// ListStorageClasses 获取 storageclass 列表
func (h *Handler) ListStorageClasses(c *gin.Context) {
	scs, err := h.k8sService.ListStorageClasses(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": scs})
}

// ListDaemonSets 获取 daemonset 列表
func (h *Handler) ListDaemonSets(c *gin.Context) {
	namespace := c.Query("namespace")
	daemonsets, err := h.k8sService.ListDaemonSets(c.Request.Context(), namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": daemonsets})
}
