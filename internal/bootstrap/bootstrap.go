package bootstrap

import (
	"fmt"
	"modserv-shim/api/server"
	"modserv-shim/internal/config"
	"modserv-shim/internal/core/eventbus"
	"modserv-shim/internal/core/eventbus/subscriptions"
	"modserv-shim/internal/core/orchestrator"
	"modserv-shim/internal/core/pipeline"
	"modserv-shim/internal/core/shimlet"
	_ "modserv-shim/internal/core/shimlet/shimlets"
	"modserv-shim/internal/core/statemanager"
	"modserv-shim/internal/core/tracer"
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

	// shimlet registry already initialed from init()
	shimReg := shimlet.Registry
	pipeReg := pipeline.Registry

	// TODO init stateManager(FSM)
	stateMgr := statemanager.New()

	// init eventbus (default use asaskevich impl)
	eventbusInstance := eventbus.NewAsaskevichEventBus()

	// init eventbus subscriptions
	subscriptions.Setup(eventbusInstance, stateMgr)

	// 初始化全局Tracer单例
	statusTracer := tracer.New(eventbusInstance)
	shim, _ := shimReg.GetSingleton(cfg.CurrentShimlet)

	// Trace 已部署的服务
	err := statusTracer.Init(shim, 30)
	if err != nil {
		return err
	}

	log.Info("Global tracer initialized")

	// init orchestrator
	orchestrator.GlobalOrchestrator = orchestrator.NewOrchestrator(shimReg, pipeReg, eventbusInstance, statusTracer, stateMgr)

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
