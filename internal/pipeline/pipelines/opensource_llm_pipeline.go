package pipelines

import (
	dto "modserv-shim/internal/dto/deploy"
	"modserv-shim/internal/pipeline"
)

type OpensourceLLMPipeline struct {
}

func init() {
	pipeline.Registry.AutoRegister(&OpensourceLLMPipeline{})
}

func (*OpensourceLLMPipeline) Apply(spec dto.DeploySpec) (resourceId string, err error) {
	return "", err
}
func (*OpensourceLLMPipeline) ID() (name string)   { return "" }
func (*OpensourceLLMPipeline) Description() string { return "" }
