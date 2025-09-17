package shimdrive

import (
	dto "modserv-shim/internal/dto/deploy"
	"modserv-shim/internal/shimreg"
)

type ShimDrive struct {
	GlobalShimlet shimreg.Shimlet
}

func (d *ShimDrive) deploy(depSpec dto.DeploySpec) {

	// TODO 渲染部署文件

	// TODO track状态

	// TODO 调用对应 shimlet 执行部署操作

	// TODO 颁发 serviceID 并暴露 endpoint
}
