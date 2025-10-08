package state

import (
	dto "modserv-shim/internal/dto/deploy"
	"sync"
)

type ServiceState string

const (
	Pending ServiceState = "pending"
	Running ServiceState = "running"
	Failed  ServiceState = "failed"
)

// Manager 是 StateManager 的简单内存实现
type Manager struct {
	specMap map[string]*dto.DeploySpec
	mu      sync.RWMutex
}

// New 创建一个新的 StateManager 实例
func New() *Manager {
	return &Manager{
		specMap: make(map[string]*dto.DeploySpec),
	}
}

// Set 保存用户部署期望 以及 runtime shimlet 和 部署 goal set (目标集合)
func (m *Manager) Set(serviceID string, spec *dto.DeploySpec) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.specMap[serviceID]; !exists {
		m.specMap[serviceID] = spec
	}
}

// Delete 删除服务的状态记录
func (m *Manager) Delete(serviceID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.specMap, serviceID)
}

func (m *Manager) GetStatus(id string) {

	// TODO 判断goal set 所有 goals is achieved
}
