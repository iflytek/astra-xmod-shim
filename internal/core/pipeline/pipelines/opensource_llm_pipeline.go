package pipelines

import (
	"modserv-shim/internal/config"
	"modserv-shim/internal/core/pipeline"
	"modserv-shim/pkg/log"
	"path/filepath"
)

func init() {
	opensourceLLMPipeline()
}

// mapModelNameToPath maps model name to actual model path
func mapModelNameToPath(ctx *pipeline.Context) error {
	// Get model root directory from config
	modelRoot := config.Get().ModelManage.ModelRoot
	if modelRoot == "" {
		modelRoot = "/models"
		log.Info("Using default model root directory: %s", modelRoot)
	}

	// Construct full model path
	modelDir := filepath.Join(modelRoot, ctx.DeploySpec.ModelName)
	log.Info("Mapping model name to path: %s -> %s", ctx.DeploySpec.ModelName, modelDir)

	// Set mapped path to DeploySpec
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

func StartTracker(ctx *pipeline.Context) error {
	err := ctx.Tracer.Trace(ctx.ResourceId, ctx.Shimlet, 10)
	if err != nil {
		return err
	}
	return nil
}

func opensourceLLMPipeline() *pipeline.Pipeline {

	return pipeline.New("opensource_llm").
		Step(mapModelNameToPath). // Step: Map model name to path
		Step(StartTracker).
		Step(applyService).
		BuildAndRegister()
}
