package handler

import (
	"astron-xmod-shim/internal/core/orchestrator"
	dto "astron-xmod-shim/internal/dto/deploy"
	"astron-xmod-shim/pkg/log"
	"astron-xmod-shim/pkg/utils"
	"net/http"
	"time"

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

// GetServiceStatusResponse 获取服务状态响应结构体
type GetServiceStatusResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		ServiceID  string `json:"serviceId"`
		Status     string `json:"status"`   // 运行中/阻塞中/失败/初始化中/不存在/停止中
		Endpoint   string `json:"endpoint"` // openai like endpoint
		UpdateTime string `json:"updateTime"`
	} `json:"data"`
}

func DoDeploy(c *gin.Context) {
	var depSpec *dto.RequirementSpec
	if err := c.ShouldBindJSON(&depSpec); err != nil {
		log.Error("解析策略请求失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"Code":    http.StatusBadRequest,
			"Message": "无效的请求参数: " + err.Error(),
		})
		return
	}

	depSpec.ServiceId = utils.GenerateSimpleID()
	depSpec.GoalSetName = "opensource-llm-deploy"
	err := orchestrator.GlobalOrchestrator.Provision(depSpec)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    1,
			"message": "deploy submit failed: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "deploy submit success",
		"data":    map[string]string{"serviceId": depSpec.ServiceId},
	})
}

// GetServiceStatus 处理获取模型服务状态的请求
func GetServiceStatus(c *gin.Context) {
	// 从URL路径中获取serviceId
	serviceID := c.Param("serviceId")

	if serviceID == "" {
		log.Error("serviceId is required")
		response := GetServiceStatusResponse{
			Code:    1,
			Message: "serviceId is required",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	log.Info("Getting service status", "serviceID", serviceID)

	// 调用orchestrator获取服务状态
	status, err := orchestrator.GlobalOrchestrator.GetServiceStatus(serviceID)
	if err != nil {
		log.Error("Get service status failed", "error", err)
		response := GetServiceStatusResponse{
			Code:    1,
			Message: "get service status failed",
			Data: struct {
				ServiceID  string `json:"serviceId"`
				Status     string `json:"status"`
				Endpoint   string `json:"endpoint"`
				UpdateTime string `json:"updateTime"`
			}{serviceID, "", "", ""},
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// 构建OpenAI风格的endpoint（实际应该从K8s服务或配置中获取）

	// 获取当前时间
	updateTime := time.Now().Format("2006-01-02 15:04:05")

	// 返回成功响应
	response := GetServiceStatusResponse{
		Code:    0,
		Message: "success",
		Data: struct {
			ServiceID  string `json:"serviceId"`
			Status     string `json:"status"`
			Endpoint   string `json:"endpoint"`
			UpdateTime string `json:"updateTime"`
		}{serviceID, string(status.Status), status.EndPoint, updateTime},
	}
	c.JSON(http.StatusOK, response)
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

	spec := &dto.RequirementSpec{GoalSetName: "opensource-llm-delete"}

	err := orchestrator.GlobalOrchestrator.Provision(spec)
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

// UpdateService 处理更新模型服务的请求
func UpdateService(c *gin.Context) {
	// 从URL路径中获取serviceId
	serviceID := c.Param("serviceId")

	if serviceID == "" {
		log.Error("serviceId is required")
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": "serviceId is required",
		})
		return
	}

	var depSpec *dto.RequirementSpec
	if err := c.ShouldBindJSON(&depSpec); err != nil {
		log.Error("解析策略请求失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": "无效的请求参数: " + err.Error(),
		})
		return
	}

	// 使用URL中的serviceId，而不是生成新的
	depSpec.ServiceId = serviceID

	log.Info("Updating service", "serviceID", serviceID)
	depSpec.GoalSetName = "opensource-llm-deploy"
	// 复用部署逻辑进行更新
	err := orchestrator.GlobalOrchestrator.Provision(depSpec)
	if err != nil {
		log.Error("Update service failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    1,
			"message": "update submit failed: " + err.Error(),
			"data":    map[string]string{"serviceId": serviceID},
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "update submit success",
		"data":    map[string]string{"serviceId": serviceID},
	})
}
