package pipeline

import (
	dto "modserv-shim/internal/dto/deploy"
	"modserv-shim/internal/typereg"
)

var Registry = typereg.New[Pipeline]()

// Pipeline is the interface that must be implemented by all shimlets.
type Pipeline interface {
	CreateDeploySpec() (spec dto.DeploySpec, err error)
	apply(spec dto.DeploySpec) (resourceId string, err error)
	Delete(resourceId string) error
	Status(resourceId string) (status *dto.DeployStatus, err error)
	ID() (name string)
	Description() string
}
