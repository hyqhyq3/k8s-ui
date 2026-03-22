package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler HTTP 处理器
type Handler struct {
}

// NewHandler 创建处理器
func NewHandler() *Handler {
	return &Handler{}
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
