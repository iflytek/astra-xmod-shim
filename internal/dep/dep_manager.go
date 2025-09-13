package dep

import (
	"context"
	model "modserv-shim/internal/model/dep"
	"modserv-shim/internal/shim"
)

type DeployManager struct {
	shimlet       shim.Shimlet
	monitorCtxMap map[string]context.Context
}

func (d *DeployManager) deploy(depSpec model.DeploySpec) {

	// TODO 渲染部署文件

	// TODO 绑定状态 goroutine sidecar

	// TODO 调用对应 shimlet 执行部署操作

	// TODO 颁发 serviceID 并暴露 endpoint
}
