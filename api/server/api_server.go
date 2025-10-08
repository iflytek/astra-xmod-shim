package server

import (
	"astron-xmod-shim/api/middleware"
	"astron-xmod-shim/api/route"
	"astron-xmod-shim/internal/config"
	"astron-xmod-shim/pkg/http"
	"astron-xmod-shim/pkg/log"

	"github.com/gin-gonic/gin"
)

// Init 启动HttpServer
func Init() error {

	gin.SetMode(gin.ReleaseMode) // 放在初始化 Engine 之前
	// 2. 后续按需获取配置（首次调用Get()时完整初始化）
	globalCfg := config.Get()
	log.Info("HTTP服务器地址端口%v", globalCfg.Server.Port)

	// 3. 初始化通用HTTP服务器
	httpServer := http.NewServer(globalCfg.Server.Port)

	// 注册业务路由
	route.RegisterRoutes(httpServer)

	// 注册日志中间件
	engine := httpServer.GetEngine()
	engine.Use(middleware.Logging())

	log.Info("HTTP服务器初始化完毕")

	// 启动服务器
	return httpServer.Run()
}
