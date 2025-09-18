package orchestrator

import (
	"modserv-shim/internal/core/pipeline"
	"modserv-shim/internal/core/shimlet"
	dto "modserv-shim/internal/dto/deploy"
)

type ExecContext struct {
	spec    dto.DeploySpec
	shimlet *shimlet.Shimlet
	pipe    *pipeline.Pipeline
	phase   dto.DeployPhase
}
