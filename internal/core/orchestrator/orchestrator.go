package orchestrator

import (
	"modserv-shim/internal/config"
	"modserv-shim/internal/core/pipeline"
	"modserv-shim/internal/core/shimlet"
	"modserv-shim/internal/core/typereg"
	dto "modserv-shim/internal/dto/deploy"
	"modserv-shim/pkg/log"
)

type Orchestrator struct {
	ShimReg *typereg.TypeReg[shimlet.Shimlet]
	//PipeReg *typereg.TypeReg[*pipeline.Pipeline]
}

func (d *Orchestrator) Provision(spec dto.DeploySpec) {

	// TODO 渲染部署文件
	_ = d.newExecCtx(spec)

	// TODO 调用对应 shimlet 执行部署操作

	// TODO track状态

	// TODO 颁发 serviceID 并暴露 endpoint
}

// TODO: [临时] 后续应根据 spec.Type 或 metadata 动态选择 pipeline
// 示例：spec.Type == "llm" → "ai-pipeline", spec.Type == "web" → "web-pipeline"
func getPipelineName() string {
	return "opensource_llm"
}

func (d *Orchestrator) newExecCtx(spec dto.DeploySpec) (ctx *ExecContext) {

	shimletId := config.Get().CurrentShimlet
	pipelineId := getPipelineName() // mock method

	runtimeShimlet, err := shimlet.Registry.GetSingleton(shimletId)
	runtimePipe := pipeline.Registry[pipelineId]

	if err != nil {
		log.Error("init exec context error", err)
		return ctx
	}

	ctx = &ExecContext{
		shimlet: runtimeShimlet,
		pipe:    runtimePipe,
		spec:    spec,
	}
	return ctx
}
