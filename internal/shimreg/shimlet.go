package shimreg

import dto "modserv-shim/internal/dto/deploy"

// Shimlet is the interface that must be implemented by all shimlets.
type Shimlet interface {
	InitWithConfig(confPath string) error
	Create(spec dto.DeploySpec) (resourceId string, err error)
	Update(spec dto.DeploySpec) (resourceId string, err error)
	Delete(resourceId string) error
	Status(resourceId string) (status *dto.DeployStatus, err error)
	ID() (name string)
	Description() string
}
