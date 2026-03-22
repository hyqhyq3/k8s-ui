package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yangqihuang/k8s-ui/internal/helm"
	"github.com/yangqihuang/k8s-ui/internal/service"
)

// HelmHandler Helm HTTP 处理器
type HelmHandler struct {
	helmService *service.HelmService
}

// NewHelmHandler 创建 Helm 处理器
func NewHelmHandler(helmService *service.HelmService) *HelmHandler {
	return &HelmHandler{helmService: helmService}
}

// ListReleases 列出 Helm release
func (h *HelmHandler) ListReleases(c *gin.Context) {
	namespace := c.Query("namespace")

	releases, err := h.helmService.ListReleases(c.Request.Context(), namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": releases})
}

// GetRelease 获取 release 详情
func (h *HelmHandler) GetRelease(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	detail, err := h.helmService.GetRelease(namespace, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": detail})
}

// GetReleaseHistory 获取 release revision 历史
func (h *HelmHandler) GetReleaseHistory(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	history, err := h.helmService.GetReleaseHistory(namespace, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": history})
}

// GetReleaseResources 获取 release 管理的资源
func (h *HelmHandler) GetReleaseResources(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	resources, err := h.helmService.GetReleaseResources(namespace, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": resources})
}

// UninstallRelease 卸载 release
func (h *HelmHandler) UninstallRelease(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	keepHistory := c.Query("keepHistory") == "true"

	if err := h.helmService.UninstallRelease(namespace, name, keepHistory); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"message": "uninstalled"}})
}

// RollbackRelease 回滚 release
func (h *HelmHandler) RollbackRelease(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	var req struct {
		Revision int `json:"revision"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	if err := h.helmService.RollbackRelease(namespace, name, req.Revision); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"message": "rolled back"}})
}

// InstallRelease 安装 release
func (h *HelmHandler) InstallRelease(c *gin.Context) {
	var req helm.InstallOptions
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数: " + err.Error()})
		return
	}

	if req.Name == "" || req.Namespace == "" || req.Chart == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name, namespace, chart 不能为空"})
		return
	}

	detail, err := h.helmService.InstallRelease(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": detail})
}

// UpgradeRelease 升级 release
func (h *HelmHandler) UpgradeRelease(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	var req helm.UpgradeOptions
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数: " + err.Error()})
		return
	}

	req.Name = name
	req.Namespace = namespace

	detail, err := h.helmService.UpgradeRelease(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": detail})
}

// ListRepos 列出 chart repo
func (h *HelmHandler) ListRepos(c *gin.Context) {
	repos, err := h.helmService.ListRepos()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": repos})
}

// AddRepo 添加 chart repo
func (h *HelmHandler) AddRepo(c *gin.Context) {
	var req struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	if req.Name == "" || req.URL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name 和 url 不能为空"})
		return
	}

	if err := h.helmService.AddRepo(req.Name, req.URL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"message": "repo added"}})
}

// RemoveRepo 删除 chart repo
func (h *HelmHandler) RemoveRepo(c *gin.Context) {
	name := c.Param("name")

	if err := h.helmService.RemoveRepo(name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"message": "repo removed"}})
}

// SearchChart 搜索 chart
func (h *HelmHandler) SearchChart(c *gin.Context) {
	repoName := c.Param("repo")
	keyword := c.Query("q")

	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "搜索关键词不能为空"})
		return
	}

	results, err := h.helmService.SearchChart(repoName, keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": results})
}

// GetChartVersions 获取 chart 可用版本列表
func (h *HelmHandler) GetChartVersions(c *gin.Context) {
	repoName := c.Param("repo")
	chartName := c.Param("chart")

	versions, err := h.helmService.GetChartVersions(repoName, chartName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": versions})
}

// Helper: parse query param to int with default
func queryInt(c *gin.Context, key string, defaultVal int) int {
	val := c.Query(key)
	if val == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return n
}
