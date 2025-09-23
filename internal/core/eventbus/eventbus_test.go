package eventbus

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// newTestEventBus 创建一个用于测试的EventBus实例
func newTestEventBus() EventBus {
	return NewAsaskevichEventBus()
}

// 测试基本的发布订阅功能 - 通过EventBus接口
func TestBasicPublishSubscribe(t *testing.T) {
	// 使用EventBus接口类型
	var bus EventBus = newTestEventBus()

	// 测试普通订阅和发布
	var receivedMsg string
	var receivedArgs []interface{}

	subscriber := func(msg string, args ...interface{}) {
		receivedMsg = msg
		receivedArgs = args
	}

	// 通过接口订阅
	err := bus.Subscribe("test-topic", subscriber)
	assert.NoError(t, err)

	// 通过接口发布
	testArgs := []interface{}{"hello", "arg1", 123, true}
	bus.Publish("test-topic", testArgs...)

	// 验证接收
	assert.Equal(t, "hello", receivedMsg)
	assert.Len(t, receivedArgs, 3) // 现在应该有3个参数："arg1", 123, true
	assert.Equal(t, "arg1", receivedArgs[0])
	assert.Equal(t, 123, receivedArgs[1])
	assert.Equal(t, true, receivedArgs[2])

	// 通过接口取消订阅
	err = bus.Unsubscribe("test-topic", subscriber)
	assert.NoError(t, err)

	// 再次发布，不应该收到
	receivedMsg = ""
	receivedArgs = nil
	bus.Publish("test-topic", "should-not-receive")
	assert.Equal(t, "", receivedMsg)
	assert.Nil(t, receivedArgs)
}

// 测试多订阅者 - 通过EventBus接口
func TestMultipleSubscribers(t *testing.T) {
	// 使用EventBus接口类型
	var bus EventBus = newTestEventBus()

	var wg sync.WaitGroup
	receivedCount := 0
	mutex := &sync.Mutex{}

	subscriber1 := func(msg string) {
		defer wg.Done()
		mutex.Lock()
		receivedCount++
		mutex.Unlock()
	}

	subscriber2 := func(msg string) {
		defer wg.Done()
		mutex.Lock()
		receivedCount++
		mutex.Unlock()
	}

	// 通过接口订阅
	bus.Subscribe("multi-topic", subscriber1)
	bus.Subscribe("multi-topic", subscriber2)

	// 通过接口发布
	wg.Add(2)
	bus.Publish("multi-topic", "test")
	wg.Wait()

	// 验证两个订阅者都收到了消息
	assert.Equal(t, 2, receivedCount)
}

// 测试订阅者参数类型错误 - 通过EventBus接口
func TestInvalidSubscriberType(t *testing.T) {
	// 使用EventBus接口类型
	var bus EventBus = newTestEventBus()

	// 非函数类型的订阅者
	err := bus.Subscribe("test-topic", "not-a-function")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "subscriber must be a function")

	// nil订阅者
	err = bus.Subscribe("test-topic", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "subscriber must be a function")
}

// 测试取消订阅参数类型错误 - 通过EventBus接口
func TestInvalidUnsubscriberType(t *testing.T) {
	// 使用EventBus接口类型
	var bus EventBus = newTestEventBus()

	// 非函数类型的取消订阅者
	err := bus.Unsubscribe("test-topic", "not-a-function")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsubscriber must be a function")

	// nil取消订阅者
	err = bus.Unsubscribe("test-topic", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsubscriber must be a function")
}

// 测试全局单例 - 通过EventBus接口
func TestGlobalEventBus(t *testing.T) {
	// 重置全局单例（测试环境下）
	globalEventBus = nil
	initOnce = sync.Once{}

	// 创建测试用的EventBus
	var bus EventBus = newTestEventBus()

	// 初始化全局单例
	InitGlobalEventBus(bus)

	// 获取全局单例 - 返回EventBus接口
	globalBus := GetGlobalEventBus()
	assert.NotNil(t, globalBus)
	assert.Equal(t, bus, globalBus)

	// 测试全局单例的发布订阅 - 通过接口
	var received bool
	subscriber := func() {
		received = true
	}

	globalBus.Subscribe("global-topic", subscriber)
	globalBus.Publish("global-topic")
	assert.True(t, received)

	// 测试多次初始化只生效一次
	var newBus EventBus = newTestEventBus()
	InitGlobalEventBus(newBus)
	assert.Equal(t, bus, GetGlobalEventBus()) // 应该还是原来的实例
}

// 测试并发安全性 - 通过EventBus接口
func TestConcurrencySafety(t *testing.T) {
	// 使用EventBus接口类型
	var bus EventBus = newTestEventBus()

	var wg sync.WaitGroup
	var mutex sync.Mutex
	var receivedMessages []string

	// 订阅者
	subscriber := func(msg string) {
		mutex.Lock()
		receivedMessages = append(receivedMessages, msg)
		mutex.Unlock()
	}

	// 通过接口订阅
	bus.Subscribe("concurrent-topic", subscriber)

	// 并发通过接口发布
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			bus.Publish("concurrent-topic", "message-", id)
		}(i)
	}

	wg.Wait()
	// 验证所有消息都被接收
	assert.Equal(t, 100, len(receivedMessages))
}

// 测试性能 - 通过EventBus接口
func BenchmarkEventBusPublishSubscribe(b *testing.B) {
	// 使用EventBus接口类型
	var bus EventBus = newTestEventBus()

	// 简单的订阅者函数
	subscriber := func(msg string) {}
	bus.Subscribe("bench-topic", subscriber)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bus.Publish("bench-topic", "benchmark-message")
	}
}

// 测试不同类型的参数传递 - 通过EventBus接口
func TestDifferentParameterTypes(t *testing.T) {
	// 使用EventBus接口类型
	var bus EventBus = newTestEventBus()

	// 测试不同类型的参数
	type TestStruct struct {
		Name  string
		Value int
	}

	var receivedString string
	var receivedInt int
	var receivedBool bool
	var receivedStruct TestStruct

	subscriber := func(s string, i int, b bool, ts TestStruct) {
		receivedString = s
		receivedInt = i
		receivedBool = b
		receivedStruct = ts
	}

	// 通过接口订阅
	bus.Subscribe("types-topic", subscriber)

	testStruct := TestStruct{Name: "test", Value: 42}
	// 通过接口发布
	bus.Publish("types-topic", "string-value", 123, true, testStruct)

	assert.Equal(t, "string-value", receivedString)
	assert.Equal(t, 123, receivedInt)
	assert.Equal(t, true, receivedBool)
	assert.Equal(t, testStruct, receivedStruct)
}
