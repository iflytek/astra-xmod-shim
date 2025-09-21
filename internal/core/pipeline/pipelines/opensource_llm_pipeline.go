package pipelines

import (
	"modserv-shim/internal/config"
	"modserv-shim/internal/core/pipeline"
	"modserv-shim/pkg/log"
	"modserv-shim/pkg/utils"
	"path/filepath"
)

func init() {
	opensourceLLMPipeline()
}
func generateServiceId(ctx *pipeline.Context) error {
	if ctx.DeploySpec.ServiceId == "" {
		ctx.DeploySpec.ServiceId = utils.GenerateSimpleID()
	}
	return nil
}

// mapModelNameToPath 将模型名称映射到实际的模型路径
func mapModelNameToPath(ctx *pipeline.Context) error {
	// 获取配置中的模型根目录
	modelRoot := config.Get().ModelManage.ModelRoot
	if modelRoot == "" {
		modelRoot = "/models"
		log.Info("使用默认模型根目录: %s", modelRoot)
	}
	
	// 构建完整的模型路径
	modelDir := filepath.Join(modelRoot, ctx.DeploySpec.ModelName)
	log.Info("根据模型名称映射到路径: %s -> %s", ctx.DeploySpec.ModelName, modelDir)
	
	// 将映射后的路径设置到DeploySpec
	ctx.DeploySpec.ModelFileDir = modelDir
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
		Step(mapModelNameToPath). // 添加模型名称到路径的映射步骤
		Step(applyService).
		Step(exposeService).
		BuildAndRegister()
}