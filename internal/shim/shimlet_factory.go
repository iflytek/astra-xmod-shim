// 实现ShimletFactory，负责创建所有类型的Shimlet实例
package shim

import (
	"fmt"
)

// ShimletFactory 万能工厂，创建所有类型的Shimlet
type ShimletFactory struct{}

// CreateShimlet 根据类型创建Shimlet实例
func (f *ShimletFactory) CreateShimlet(shimType string) (Shimlet, error) {
	switch shimType {
	case "k8s":
		return nil, nil
	case "docker":
		return nil, nil
	case "shimmy":
		return nil, nil
	default:
		return nil, fmt.Errorf("不支持的类型：%s", shimType)
	}
}
