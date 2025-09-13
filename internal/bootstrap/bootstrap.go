package bootstrap

import (
	"fmt"
	cfgUtil "modserv-shim/internal/cfg"
	"modserv-shim/internal/server"
	"modserv-shim/pkg/log"
	"sync"
)

var (
	wg sync.WaitGroup
)

func Init(configPath string) error {
	// TODO bootstrap serval steps impls
	// 1. 加载配置（日志未初始化，用fmt输出错误）
	cfgUtil.SetConfigPath(configPath)
	cfg, err := cfgUtil.Get()
	if err != nil {
		return fmt.Errorf("配置文件加载失败: %w", err) // 此时日志未就绪，返回错误由上层处理
	}
	fmt.Println("配置文件加载完成")

	// 2. 用配置初始化日志系统（日志配置来自第一步加载的cfg）
	if err := log.Init(&cfg.Log); err != nil {
		return fmt.Errorf("日志初始化失败: %w", err) // 日志初始化失败，无法使用log输出
	}
	log.Info("日志系统初始化完成", "配置", cfg.Log)

	// 3. 初始化HTTP服务器

	// 注册gin的日志中间件
	//engine.Use(middleware.Logging())

	// 6. 初始化 HTTP Server
	if err := server.Init(); err != nil {
		return fmt.Errorf("HTTP Server初始化失败: %w", err)
	}

	return nil
}

// registerShutdownHook
func registerShutdownHook() {
	// TODO shutdown hook impl

}

func WaitForShutDown() {
	// TODO graceful shutdown logics
}
