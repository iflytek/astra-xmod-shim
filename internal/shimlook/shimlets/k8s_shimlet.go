package shimlets

import (
	dto "modserv-shim/internal/dto/deploy"
	"modserv-shim/internal/shimlook"
)

// 编译时检查 确保实现 shimlet 接口
var _ shimlook.Shimlet = (*k8sShimlet)(nil)

type k8sShimlet struct {
}

func (k k8sShimlet) InitWithConfig(confPath string) error {
	return nil
}

func (k k8sShimlet) ValidateConfig() {}
func (k k8sShimlet) Create(spec dto.DeploySpec) (resourceId string, err error) {
	return "", err
}
func (k k8sShimlet) Update(spec dto.DeploySpec) (resourceId string, err error) {
	return "", err
}
func (k k8sShimlet) Delete(resourceId string) (err error) { return nil }
func (k k8sShimlet) Status(resourceId string) (status *dto.DeployStatus, err error) {
	return nil, err
}
