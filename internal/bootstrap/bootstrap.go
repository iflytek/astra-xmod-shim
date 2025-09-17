package bootstrap

import (
	"fmt"
	"modserv-shim/api/server"
	"modserv-shim/internal/config"
	"modserv-shim/internal/engine/shimdrive"
	"modserv-shim/internal/shimreg"
	_ "modserv-shim/internal/shimreg/shimlets"
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

	// init shimReg
	runtimeShimlet := shimreg.NewUninitialized(cfg.CurrentShimlet)
	err := runtimeShimlet.InitWithConfig(cfg.Shimlets[runtimeShimlet.ID()].ConfigPath)
	if err != nil {
		log.Error("shimlet init failed", err)
	}

	// TODO 初始化 pipeLook

	// TODO 初始化 shimDrive
	_ = &shimdrive.ShimDrive{GlobalShimlet: runtimeShimlet}

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
