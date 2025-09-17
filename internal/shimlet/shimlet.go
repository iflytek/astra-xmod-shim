package shimlet

import (
	dto "modserv-shim/internal/dto/deploy"
	"modserv-shim/internal/typereg"
)

var Registry = typereg.New[Shimlet]()

// Shimlet is the interface that must be implemented by all shimlets.
type Shimlet interface {
	InitWithConfig(confPath string) error
	Apply(spec dto.DeploySpec) (resourceId string, err error)
	Delete(resourceId string) error
	Status(resourceId string) (status *dto.DeployStatus, err error)
	ID() (name string)
	Description() string
}
