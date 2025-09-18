package bootstrap

import (
	"fmt"
	"modserv-shim/api/server"
	"modserv-shim/internal/config"
	"modserv-shim/internal/core/orchestrator"
	"modserv-shim/internal/core/pipeline"
	"modserv-shim/internal/core/shimlet"
	"modserv-shim/internal/engine/shimdrive"
	_ "modserv-shim/internal/shimlet/shimlets"
	"modserv-shim/pkg/log"
	"sync"
)

var (
	wg sync.WaitGroup
)

func Init(configPath string) error {
	// init config
	config.SetConfigPath(configPath)
	cfg := config.Get()

	// init log
	if err := log.Init(&cfg.Log); err != nil {
		return fmt.Errorf("log configured error: %w", err) // 日志初始化失败，无法使用log输出
	}
	log.Info("log configured", "cfg: ", cfg.Log)

	// registry already initialed from init()
	shimReg := shimlet.Registry
	pipeReg := pipeline.Registry

	// TODO 初始化 shimDrive
	_ = &orchestrator.Orchestrator{ShimReg: shimReg, PipeReg: pipeReg}

	// TODO 初始化 stateTrack

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
