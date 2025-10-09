package spec

import (
	dto "astron-xmod-shim/internal/dto/deploy"
)

// Store 是 StateManager 的简单内存实现
type Store interface {
	Set(serviceID string, spec *dto.RequirementSpec)
	Delete(serviceID string)
	Get(serviceID string) *dto.RequirementSpec
}
