package pipelines

import (
	"modserv-shim/internal/core/pipeline"
	"modserv-shim/pkg/log"
	"modserv-shim/pkg/utils"
)

func generateServiceId(ctx *pipeline.Context) error {
	if ctx.DeploySpec.ServiceId == "" {
		ctx.DeploySpec.ServiceId = utils.GenerateSimpleID()
	}
	return nil
}

func applyService(ctx *pipeline.Context) error {
	resourceId, err := ctx.Shimlet.Apply(ctx.DeploySpec)
	if err != nil {
		return err
	}
	ctx.ResourceId = resourceId
	return nil
}

func Track(ctx *pipeline.Context) error {

	// TODO 部署之前 调用 tracer 跟踪状态

	return nil
}

func exposeService(ctx *pipeline.Context) error {
	// 从上下文中获取resourceId
	resourceId, ok := ctx.Get("resourceId").(string)
	if !ok || resourceId == "" {
		log.Warn("No resourceId found in context, cannot expose service properly")
	}

	// TODO 拼接 endpoint

	return nil
}

func opensourceLLMPipeline() *pipeline.Pipeline {

	return pipeline.New("opensource_llm").
		Step(generateServiceId).
		Step(applyService).
		Step(exposeService).
		BuildAndRegister()
}
