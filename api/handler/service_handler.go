package handler

import (
	"modserv-shim/internal/core/orchestrator"
	model "modserv-shim/internal/dto/deploy"
	"modserv-shim/pkg/log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// DeleteServiceResponse 删除服务响应结构体
type DeleteServiceResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		ServiceID string `json:"serviceId"`
	} `json:"data"`
}
func DoDeploy(c *gin.Context) {
	var depSpec *model.DeploySpec
	if err := c.ShouldBindJSON(&depSpec); err != nil {
		log.Error("解析策略请求失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"Code":    http.StatusBadRequest,
			"Message": "无效的请求参数: " + err.Error(),
		})
		return
	}

	err := orchestrator.GlobalOrchestrator.Provision(depSpec)
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}
// DeleteService 处理删除模型服务的请求
func DeleteService(c *gin.Context) {
	// 从URL路径中获取serviceId
	serviceID := c.Param("serviceId")

	if serviceID == "" {
		log.Error("serviceId is required")
		response := DeleteServiceResponse{
			Code:    1,
			Message: "serviceId is required",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	log.Info("Deleting service", "serviceID", serviceID)

	// 调用orchestrator删除服务
	err := orchestrator.GlobalOrchestrator.DeleteService(serviceID)
	if err != nil {
		log.Error("Delete service failed", "error", err)
		response := DeleteServiceResponse{
			Code:    1,
			Message: "delete submit failed",
			Data: struct {
				ServiceID string `json:"serviceId"`
			}{serviceID},
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// 返回成功响应
	response := DeleteServiceResponse{
		Code:    0,
		Message: "delete submit success",
		Data: struct {
			ServiceID string `json:"serviceId"`
		}{serviceID},
	}
	c.JSON(http.StatusOK, response)
}