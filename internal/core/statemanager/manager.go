// pkg/statemanager/state_manager.go
package statemanager

import (
	dto "modserv-shim/internal/dto/deploy"
	"modserv-shim/internal/dto/eventbus"
	"sync"
)

// StateManager holds the current phase of each service
type StateManager struct {
	states map[string]dto.DeployPhase
	mutex  sync.RWMutex
}

// New creates a new StateManager
func New() *StateManager {
	return &StateManager{
		states: make(map[string]dto.DeployPhase),
	}
}

// UpdateStatus handles a status update event from EventBus
// This is the ONLY method you need
func (sm *StateManager) UpdateStatus(event *eventbus.ServiceEvent) {
	sm.mutex.Lock()
	sm.states[event.ServiceID] = event.To
	sm.mutex.Unlock()
}

// Get returns the current phase of a service

func (sm *StateManager) Get(serviceID string) dto.DeployPhase {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	if phase, exists := sm.states[serviceID]; exists {
		return phase
	}

	return dto.PhaseUnknown // ✅ 返回 unknown，而不是空字符串
}
