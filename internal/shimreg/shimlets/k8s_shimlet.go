package shimlets

import (
	dto "modserv-shim/internal/dto/deploy"
	"modserv-shim/internal/shimreg"
)

// 编译时检查 确保实现 shimlet 接口
var _ shimreg.Shimlet = (*K8sShimlet)(nil)

func init() {
	shimreg.AutoRegister[*K8sShimlet]()
}

type K8sShimlet struct {
}

func (k *K8sShimlet) ID() string {
	return "k8s"
}
func (k *K8sShimlet) InitWithConfig(confPath string) error {
	return nil
}

func (k *K8sShimlet) Create(spec dto.DeploySpec) (resourceId string, err error) {
	return "", err
}
func (k *K8sShimlet) Update(spec dto.DeploySpec) (resourceId string, err error) {
	return "", err
}
func (k *K8sShimlet) Delete(resourceId string) (err error) { return nil }
func (k *K8sShimlet) Status(resourceId string) (status *dto.DeployStatus, err error) {
	return nil, err
}

func (k *K8sShimlet) Description() string {
	return "k8s shimlet"
}
