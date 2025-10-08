package spec

import (
	dto "astron-xmod-shim/internal/dto/deploy"
)

// MemoryStore 是 Store 的简单内存实现
type MemoryStore struct {
	specMap map[string]*dto.DeploySpec
}

// NewMemoryStore 创建一个新的 StateManager 实例
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		specMap: make(map[string]*dto.DeploySpec),
	}
}

// Set 保存用户部署期望 以及 runtime shimlet 和 部署 goal set (目标集合)
func (m *MemoryStore) Set(serviceID string, spec *dto.DeploySpec) {
	if _, exists := m.specMap[serviceID]; !exists {
		m.specMap[serviceID] = spec
	}
}

func (m *MemoryStore) Get(serviceID string) *dto.DeploySpec {
	return m.specMap[serviceID]
}

// Delete 删除服务的状态记录
func (m *MemoryStore) Delete(serviceID string) {
	delete(m.specMap, serviceID)
}

func (m *MemoryStore) GetStatus(id string) {

	// TODO 判断goal set 所有 goals is achieved
}
