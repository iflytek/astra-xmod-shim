package bootstrap

import (
	"fmt"
	"modserv-shim/api/server"
	cfgUtil "modserv-shim/internal/cfg"
	deploy "modserv-shim/internal/dep"
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
		return fmt.Errorf("cfg load err: %w", err) // 此时日志未就绪，返回错误由上层处理
	}

	// 2. 日志初始化
	if err := log.Init(&cfg.Log); err != nil {
		return fmt.Errorf("log configured error: %w", err) // 日志初始化失败，无法使用log输出
	}
	log.Info("log configured", "cfg: ", cfg.Log)

	// 初始化 template manager (预计淘汰掉)

	// TODO 初始化 shimDrive
	// TODO 判断初始化 shimLook

	depMgr := &deploy.DeployManager{}

	// TODO 初始化 EventBus

	// TODO 初始化 state manager

	// TODO 初始化 配置指定的 shimlet

	// TODO

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
