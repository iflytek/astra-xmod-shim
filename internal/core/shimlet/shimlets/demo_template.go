package shimlets

import (
	"modserv-shim/internal/core/shimlet"
	dto "modserv-shim/internal/dto/deploy"
)

// Compile-time check to ensure shimlet interface is implemented
var _ shimlet.Shimlet = (*TemplateShimlet)(nil)

type TemplateShimlet struct {
}

func (k *TemplateShimlet) ID() string {
	return ""
}
func (k *TemplateShimlet) InitWithConfig(confPath string) error {
	return nil
}

func (k *TemplateShimlet) Apply(spec *dto.DeploySpec) (resourceId string, err error) {
	return "", err
}
func (k *TemplateShimlet) Delete(resourceId string) (err error) { return nil }
func (k *TemplateShimlet) Status(resourceId string) (status *dto.DeployStatus, err error) {
	return nil, err
}
func (k *TemplateShimlet) Description() string {
	return "k8s shimlet"
}