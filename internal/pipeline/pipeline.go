package pipeline

import (
	dto "modserv-shim/internal/dto/deploy"
	"modserv-shim/internal/typereg"
)

var Registry = typereg.New[Pipeline]()

// Pipeline is the interface that must be implemented by all shimlets.
type Pipeline interface {
	Apply(spec dto.DeploySpec) (resourceId string, err error)
	ID() (name string)
	Description() string
}
