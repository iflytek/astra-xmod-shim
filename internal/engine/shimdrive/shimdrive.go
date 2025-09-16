package shimdrive

import (
	"modserv-shim/internal/shimlook"
)

type ShimDrive struct {
	globalShimlet shimlook.Shimlet
}

func (d *ShimDrive) deploy(depSpec dto.DeploySpec) {

	// TODO 渲染部署文件

	// TODO track状态

	// TODO 调用对应 shimlet 执行部署操作

	// TODO 颁发 serviceID 并暴露 endpoint
}
