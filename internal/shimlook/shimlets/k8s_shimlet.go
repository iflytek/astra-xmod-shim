package shimlets

import (
	dto "modserv-shim/internal/dto/deploy"
	"modserv-shim/internal/shimlook"
)

// 编译时检查 确保实现 shimlet 接口
var _ shimlook.Shimlet = (*K8sShimlet)(nil)

func init() {

}

type K8sShimlet struct {
}

func (k K8sShimlet) InitWithConfig(confPath string) (K8sShimlet *shimlook.Shimlet, err error) {
	return nil, nil
}

func (k K8sShimlet) Create(spec dto.DeploySpec) (resourceId string, err error) {
	return "", err
}
func (k K8sShimlet) Update(spec dto.DeploySpec) (resourceId string, err error) {
	return "", err
}
func (k K8sShimlet) Delete(resourceId string) (err error) { return nil }
func (k K8sShimlet) Status(resourceId string) (status *dto.DeployStatus, err error) {
	return nil, err
}
