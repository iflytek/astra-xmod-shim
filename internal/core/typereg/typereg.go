package typereg

import (
	"modserv-shim/internal/config"
	"modserv-shim/pkg/log"
	"reflect"
	"sync"
)

// TypeReg 是一个泛型注册中心
type TypeReg[T interface {
	ID() string
	InitWithConfig(confPath string) error
}] struct {
	mu                   sync.Mutex
	constructorMap       map[string]func() T
	singletonInstanceMap map[string]T
}

// New 创建一个新的 Registry
func New[T interface {
	ID() string
	InitWithConfig(confPath string) error
}]() *TypeReg[T] {
	return &TypeReg[T]{
		constructorMap:       make(map[string]func() T),
		singletonInstanceMap: make(map[string]T),
	}
}

// AutoRegister 泛型自动注册
func (r *TypeReg[T]) AutoRegister(instance T) {

	//  构造函数：返回新实例
	constructor := func() T {
		// 如果 T 是指针，直接用 reflect.New 创建
		// 如果 T 是值，也用 reflect.New 然后取 Elem
		v := reflect.New(reflect.TypeOf(instance))
		if v.Type().Elem() == reflect.TypeOf(instance) {
			return v.Elem().Interface().(T)
		}
		return v.Interface().(T)
	}

	// 3. 获取 ID 并注册
	r.mu.Lock()
	defer r.mu.Unlock()
	id := instance.ID()
	r.constructorMap[id] = constructor
}

// NewUninitialized 根据 ID 创建一个新实例
func (r *TypeReg[T]) newUninitialized(id string) T {
	if c, ok := r.constructorMap[id]; ok {
		return c()
	}
	var zero T
	return zero
}

func (r *TypeReg[T]) GetSingleton(id string) (T, error) {
	var zero T

	r.mu.Lock()
	defer r.mu.Unlock()
	singleton, exists := r.singletonInstanceMap[id]
	if !exists {
		// 实例不存在，创建并初始化
		singleton = r.newUninitialized(id)
		confPath := config.Get().Shimlets[id].ConfigPath
		if err := singleton.InitWithConfig(confPath); err != nil {
			log.Error("singleton init error: ", err)
			return zero, err // 返回零值和错误
		}
		r.singletonInstanceMap[id] = singleton
	}

	return singleton, nil
}
