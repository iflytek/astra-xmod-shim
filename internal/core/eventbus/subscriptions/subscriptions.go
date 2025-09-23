package subscriptions

import (
	"log"
	"modserv-shim/internal/core/eventbus"
	"modserv-shim/internal/core/statemanager"
)

// Setup registers all event subscriptions
func Setup(bus eventbus.EventBus, stateManager *statemanager.StateManager) {
	log.Println("🔧 Setting up event subscriptions...")

	// 1. 服务状态更新 → 更新状态机
	err := bus.Subscribe("service.status", stateManager.UpdateStatus)
	if err != nil {
		log.Printf("Failed to subscribe xxx: %v", err)
		// 不 return，继续尝试其他订阅
	}

}
