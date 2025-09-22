package route

import (
	"modserv-shim/api/handler"
	"modserv-shim/pkg/http"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册所有业务路由
func RegisterRoutes(server *http.Server) {
	// 使用修正后的GetEngine()方法获取引擎（解决引用错误）
	engine := server.GetEngine()

	// 基础API路由组
	api := engine.Group("/api")
	{
		// 版本v1路由组
		v1 := api.Group("/v1")
		{
			// 模型服务相关路由
			modserv := v1.Group("/modserv")
			{
				// 部署相关路由
				deploy := modserv.Group("/deploy")
				{
					deploy.POST("", handler.DoDeploy)
				}
				// 部署相关路由
				modList := modserv.Group("/list")
				{
					modList.GET("", handler.ListModel)
				}
				// 指标相关路由
				metrics := modserv.Group("/metrics")
				{
					metrics.GET("", func(c *gin.Context) {
						// 实现指标处理逻辑
					})
				}

				// 删除服务路由
				modserv.DELETE("/:serviceId", handler.DeleteService)
			}
		}
	}
}