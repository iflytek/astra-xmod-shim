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

func (k *TemplateShimlet) Apply(spec *dto.DeploySpec) error {
	return nil
}
func (k *TemplateShimlet) Delete(resourceId string) (err error) { return nil }
func (k *TemplateShimlet) Status(resourceId string) (status *dto.RuntimeStatus, err error) {
	return nil, err
}
func (k *TemplateShimlet) Description() string {
	return "k8s shimlet"
}

// ListDeployedServices 获取所有已部署的服务列表
// 这是一个示例实现，实际应用中需要根据具体shimlet类型实现
func (k *TemplateShimlet) ListDeployedServices() ([]string, error) {
	return []string{}, nil
}
