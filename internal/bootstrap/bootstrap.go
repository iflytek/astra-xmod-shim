package bootstrap

import (
	"astron-xmod-shim/api/server"
	"astron-xmod-shim/internal/config"
	"astron-xmod-shim/internal/core/goal"
	"astron-xmod-shim/internal/core/orchestrator"
	"astron-xmod-shim/internal/core/reconciler"
	"astron-xmod-shim/internal/core/shimlet"
	_ "astron-xmod-shim/internal/core/shimlet/shimlets"
	"astron-xmod-shim/internal/core/spec"
	"astron-xmod-shim/internal/core/workqueue"
	"astron-xmod-shim/pkg/log"
	"fmt"
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
	pipeReg := goal.Registry

	//  init specStore
	specStore := spec.NewMemoryStore()

	// init reconciler
	workerNum := 5
	workQueue := workqueue.New()

	reconciler := reconciler.NewReconciler(specStore, workerNum, workQueue)

	//  init workqueue

	// 初始化全局Tracer单例
	infraShim, _ := shimReg.GetSingleton(cfg.CurrentShimlet)

	// TODO 利用shimlet get 出服务列表
	_, _ = infraShim.ListDeployedServices()

	// init orchestrator
	orchestrator.GlobalOrchestrator = orchestrator.NewOrchestrator(shimReg, pipeReg, workQueue, specStore)

	// start reconciler
	reconciler.Start()

	// 6. 初始化 HTTP Server
	if err := server.Init(); err != nil {
		return fmt.Errorf("HTTP Server初始化失败: %w", err)
	}

	return nil
}

func WaitForShutDown() {
	// TODO graceful shutdown logics
}
