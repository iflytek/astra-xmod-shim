package shimlet

import (
	"modserv-shim/internal/core/typereg"
	dto "modserv-shim/internal/dto/deploy"
)

var Registry = typereg.New[Shimlet]()

// Shimlet is the interface that must be implemented by all shimlets.
type Shimlet interface {
	InitWithConfig(confPath string) error
	Apply(spec *dto.DeploySpec) error
	Delete(resourceId string) error
	Status(resourceId string) (status *dto.RuntimeStatus, err error)
	ID() (name string)
	Description() string
	// ListDeployedServices 获取所有已部署的服务列表
	ListDeployedServices() ([]string, error)
}
