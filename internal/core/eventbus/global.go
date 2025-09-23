package eventbus

import "sync"

// 全局单例变量
var (
	// globalEventBus 是EventBus的全局单例实例
	globalEventBus EventBus
	// 用于确保线程安全的一次性初始化
	initOnce sync.Once
)

// InitGlobalEventBus 初始化全局EventBus实例
func InitGlobalEventBus(bus EventBus) {
	initOnce.Do(func() {
		globalEventBus = bus
	})
}

// GetGlobalEventBus 获取全局EventBus单例实例
// 如果尚未初始化，会返回nil
func GetGlobalEventBus() EventBus {
	return globalEventBus
}