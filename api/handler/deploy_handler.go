package handler

import (
	model "modserv-shim/internal/dto/deploy"
	"modserv-shim/pkg/log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func DoDeploy(c *gin.Context) {
	var depSpec model.DeploySpec
	if err := c.ShouldBindJSON(&depSpec); err != nil {
		log.Error("解析策略请求失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"Code":    http.StatusBadRequest,
			"Message": "无效的请求参数: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}
