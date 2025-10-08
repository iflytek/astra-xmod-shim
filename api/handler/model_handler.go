package handler

import (
	"astron-xmod-shim/internal/config"
	"astron-xmod-shim/pkg/log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// ModelInfo 模型信息结构体
type ModelInfo struct {
	ModelName string `json:"modelName"`
	ModelPath string `json:"modelPath"`
}

// ModelListResponse 模型列表响应结构体
type ModelListResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    []ModelInfo `json:"data"`
}

func ListModel(c *gin.Context) {
	// 获取全局配置
	conf := config.Get()

	// 从配置中获取模型根目录
	modelsRootDir := conf.ModelManage.ModelRoot
	if modelsRootDir == "" {
		modelsRootDir = "/models"
		log.Info("使用默认模型根目录: %s", modelsRootDir)
	}

	log.Info("开始从模型根目录 %s 获取模型列表", modelsRootDir)

	// 检查目录是否存在并读取内容
	entries, err := os.ReadDir(modelsRootDir)
	if err != nil {
		log.Error("读取模型目录失败: %v", err)
		c.JSON(http.StatusInternalServerError, ModelListResponse{
			Code:    http.StatusInternalServerError,
			Message: "读取模型目录失败",
			Data:    []ModelInfo{},
		})
		return
	}

	// 收集模型信息（只考虑目录）
	models := make([]ModelInfo, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			models = append(models, ModelInfo{
				ModelName: entry.Name(),
				ModelPath: filepath.Join(modelsRootDir, entry.Name()),
			})
		}
	}

	log.Info("成功获取到 %d 个模型", len(models))

	// 返回模型列表
	c.JSON(http.StatusOK, ModelListResponse{
		Code:    0,
		Message: "success",
		Data:    models,
	})
}
