package pipelines

import "modserv-shim/internal/pipeline"

type OpensourceLLMPipeline struct {
}

func init() {
	pipeline.Registry.AutoRegister(&OpensourceLLMPipeline{})
}
