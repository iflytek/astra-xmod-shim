package typereg

import (
	"fmt"
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
// 支持注册指针类型实例（如 &K8sShimlet{}）和值类型实例（如 K8sShimlet{}）
// 核心修复：正确通过反射创建目标类型的非 nil 实例
func (r *TypeReg[T]) AutoRegister(instance T) {
	// 1. 获取注册实例的原始类型（如 *K8sShimlet 或 K8sShimlet）
	instanceType := reflect.TypeOf(instance)

	// 2. 定义构造函数：返回 T 类型的新实例（核心修复逻辑）
	constructor := func() T {
		var newInstanceVal reflect.Value // 反射值，用于存储创建的实例

		// 分支1：如果注册的是「指针类型」（如 &K8sShimlet{} → *K8sShimlet）
		if instanceType.Kind() == reflect.Ptr {
			// 取指针指向的「元素类型」（如 *K8sShimlet → K8sShimlet）
			elemType := instanceType.Elem()
			// 创建元素类型的实例，并取地址 → 得到目标指针类型（如 &K8sShimlet{}）
			// 此时 newInstanceVal 是 *K8sShimlet 类型，值为非 nil
			newInstanceVal = reflect.New(elemType)
		} else {
			// 分支2：如果注册的是「值类型」（如 K8sShimlet{}）
			// 先创建值类型的指针（如 *K8sShimlet），再取 Elem 得到值类型实例
			newInstanceVal = reflect.New(instanceType).Elem()
		}

		// 3. 将反射值转换为 T 类型并返回（必然成功，因为基于注册实例推导）
		return newInstanceVal.Interface().(T)
	}

	// 4. 加锁注册构造函数（原有逻辑不变，确保线程安全）
	r.mu.Lock()
	defer r.mu.Unlock()
	// 通过注册实例的 ID() 方法获取唯一标识，绑定构造函数
	id := instance.ID()
	r.constructorMap[id] = constructor

	// 可选：打印注册日志，方便调试
	fmt.Println("AutoRegister success: type=%s, id=%s", instanceType.String(), id)
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
