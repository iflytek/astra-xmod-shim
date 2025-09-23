// Package eventbus implements EventBus using github.com/asaskevich/EventBus.
package eventbus

import (
	"fmt"
	"reflect"

	asaskevich "github.com/asaskevich/EventBus"
)

// AsaskevichEventBus 是 EventBus 接口的具体实现，基于 asaskevich/EventBus
type AsaskevichEventBus struct {
	bus asaskevich.Bus
}

// NewAsaskevichEventBus 创建一个新的基于 asaskevich/EventBus 的实例
func NewAsaskevichEventBus() EventBus {
	return &AsaskevichEventBus{
		bus: asaskevich.New(),
	}
}

// Publish 发布事件
func (e *AsaskevichEventBus) Publish(topic string, args ...interface{}) {
	e.bus.Publish(topic, args...)
}

// Subscribe 订阅事件
func (e *AsaskevichEventBus) Subscribe(topic string, fn interface{}) error {
	if !isValidFunc(fn) {
		return fmt.Errorf("eventbus: subscriber must be a function, got %T", fn)
	}
	return e.bus.Subscribe(topic, fn)
}

// Unsubscribe 取消订阅
func (e *AsaskevichEventBus) Unsubscribe(topic string, fn interface{}) error {
	if !isValidFunc(fn) {
		return fmt.Errorf("eventbus: unsubscriber must be a function, got %T", fn)
	}
	return e.bus.Unsubscribe(topic, fn)
}

// isValidFunc 检查是否为函数类型
func isValidFunc(fn interface{}) bool {
	return fn != nil && reflect.TypeOf(fn).Kind() == reflect.Func
}
