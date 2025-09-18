package pipelines

import "modserv-shim/internal/core/pipeline"

func renderTemplate(ctx *pipeline.Context) error {

	return nil
}

func apply(ctx *pipeline.Context) error {

	return nil
}

func exposeService(ctx *pipeline.Context) error {
	return nil
}

func opensourceLLMPipeline() *pipeline.Pipeline {

	return pipeline.New("opensource_llm").
		Step(renderTemplate).
		Step(apply).
		Step(exposeService).
		BuildAndRegister()
}
