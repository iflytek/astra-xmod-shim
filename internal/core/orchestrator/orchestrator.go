package orchestrator

import (
	"fmt"
	"modserv-shim/internal/config"
	"modserv-shim/internal/core/pipeline"
	"modserv-shim/internal/core/shimlet"
	"modserv-shim/internal/core/typereg"
	dto "modserv-shim/internal/dto/deploy"
	"modserv-shim/pkg/log"
)

type Orchestrator struct {
	ShimReg *typereg.TypeReg[shimlet.Shimlet]
	PipeReg map[string]*pipeline.Pipeline
}

var GlobalOrchestrator *Orchestrator

func (d *Orchestrator) Provision(spec *dto.DeploySpec) error {

	runtimePipe := pipeline.Registry[getPipelineName()]
	if runtimePipe == nil {
		return fmt.Errorf("pipeline %s not found", getPipelineName())
	}

	// TODO 调用对应 shimlet 执行部署操作
	currentShimletId := config.Get().CurrentShimlet
	runtimeShimlet, err := d.ShimReg.GetSingleton(currentShimletId)
	if err != nil {
		log.Error("get runtime shimlet error", err)
		return err
	}
	pipeCtx := &pipeline.Context{
		Shimlet:    runtimeShimlet,
		DeploySpec: spec,
		Data:       make(map[string]any),
	}
	// 4. 执行pipeline
	if err := runtimePipe.Execute(pipeCtx); err != nil {
		log.Error("pipeline execution failed", err)
		return err
	}

	// TODO track状态

	// TODO 颁发 serviceID 并暴露 endpoint
	return nil
}

// TODO: [临时] 后续应根据 spec.Type 或 metadata 动态选择 pipeline
// 示例：spec.Type == "llm" → "ai-pipeline", spec.Type == "web" → "web-pipeline"
func getPipelineName() string {
	return "opensource_llm"
}
