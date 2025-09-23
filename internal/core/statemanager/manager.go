// pkg/statemanager/state_manager.go
package statemanager

import (
	dto "modserv-shim/internal/dto/deploy"
	"modserv-shim/internal/dto/eventbus"
	"sync"
)

// StateManager holds the current phase of each service
type StateManager struct {
	states map[string]*dto.RuntimeStatus
	mutex  sync.RWMutex
}

// New creates a new StateManager
func New() *StateManager {
	return &StateManager{
		states: make(map[string]*dto.RuntimeStatus),
	}
}

// UpdateStatus handles a status update event from EventBus
// This is the ONLY method you need
func (sm *StateManager) UpdateStatus(event *eventbus.ServiceEvent) {
	sm.mutex.Lock()
	sm.states[event.ServiceID] = &dto.RuntimeStatus{Status: event.To, EndPoint: event.EndPoint}
	sm.mutex.Unlock()
}

// Get returns the current phase of a service

func (sm *StateManager) GetStatus(serviceID string) *dto.RuntimeStatus {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	if status, exists := sm.states[serviceID]; exists {
		return status
	}

	return &dto.RuntimeStatus{Status: dto.PhaseUnknown} // 返回 unknown，而不是空字符串
}
