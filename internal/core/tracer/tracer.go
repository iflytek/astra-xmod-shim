package tracer

import (
	"context"
	"modserv-shim/internal/core/shimlet"
	"modserv-shim/pkg/log"
	"sync"
	"time"
)

// 全局Tracer单例
var (
	globalTracer *Tracer
	once         sync.Once
)

// GetGlobalTracer 获取全局Tracer单例实例
func GetGlobalTracer() *Tracer {
	once.Do(func() {
		globalTracer = NewTracer()
	})
	return globalTracer
}

// Trace 静态方法，直接使用全局Tracer单例跟踪服务
func Trace(serviceID string, shim shimlet.Shimlet, interval time.Duration) error {
	return GetGlobalTracer().Trace(serviceID, shim, interval)
}

// Stop 静态方法，停止指定服务的跟踪
func Stop(serviceID string) {
	GetGlobalTracer().Stop(serviceID)
}

// StopAll 静态方法，停止所有服务的跟踪
func StopAll() {
	GetGlobalTracer().StopAll()
}

// Tracer Lightweight status tracker
// Directly maintains a map of goroutine contexts without requiring an additional tracker layer

type Tracer struct {
	// Tracking task map: serviceID -> task control information
	tasks map[string]*taskControl
	mu    sync.RWMutex // Read-write lock to protect tasks
}

// taskControl Encapsulates control information for a single tracking task
// This is an internal struct and does not need to be exposed externally

type taskControl struct {
	serviceID string             // Service ID
	shimlet   shimlet.Shimlet    // Shimlet instance
	cancel    context.CancelFunc // Goroutine cancellation function
}

// NewTracer Creates a new lightweight Tracer instance

func NewTracer() *Tracer {
	return &Tracer{
		tasks: make(map[string]*taskControl),
	}
}

// Trace Method directly starts a scheduled goroutine task
// serviceID: Service instance ID
// shim: Shimlet instance to call
// interval: Status check interval time

func (t *Tracer) Trace(serviceID string, shim shimlet.Shimlet, interval time.Duration) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// 检查服务是否已经在跟踪中
	if _, exists := t.tasks[serviceID]; exists {
		log.Warn("Service is already being tracked: %s", serviceID)
		return nil
	}

	// 创建协程上下文和取消函数
	ctx, cancel := context.WithCancel(context.Background())

	// 保存任务控制信息
	t.tasks[serviceID] = &taskControl{
		serviceID: serviceID,
		shimlet:   shim,
		cancel:    cancel,
	}

	// 启动跟踪协程
	go t.trackService(ctx, serviceID, shim, interval)

	log.Info("Started tracking service: %s with shimlet: %s, interval: %v",
		serviceID, shim.ID(), interval)
	return nil
}

// Goroutine function for tracking service, implemented directly in Tracer
func (t *Tracer) trackService(ctx context.Context, serviceID string, shim shimlet.Shimlet, interval time.Duration) {
	// 确保在协程退出时清理资源
	defer func() {
		// 从任务映射中移除
		t.mu.Lock()
		delete(t.tasks, serviceID)
		t.mu.Unlock()

		log.Info("Stopped tracking service: %s", serviceID)
	}()

	// 创建定时器
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// 立即执行一次状态检查
	t.checkServiceStatus(serviceID, shim)

	// 循环执行定时检查
	for {
		select {
		case <-ctx.Done():
			// 收到取消信号，退出协程
			return
		case <-ticker.C:
			// 定时器触发，执行状态检查
			t.checkServiceStatus(serviceID, shim)
		}
	}
}

// Performs service status check
func (t *Tracer) checkServiceStatus(serviceID string, shim shimlet.Shimlet) {
	// 调用传入的shimlet方法获取服务状态
	status, err := shim.Status(serviceID)
	if err != nil {
		log.Error("Failed to get status for service %s: %v", serviceID, err)
		return
	}

	// 处理获取到的状态信息
	log.Debug("Service %s status: %+v", serviceID, status)

	// 这里可以根据实际需求添加状态处理逻辑
}

// Stop 停止指定服务的跟踪
func (t *Tracer) Stop(serviceID string) {
	t.mu.RLock()
	task, exists := t.tasks[serviceID]
	t.mu.RUnlock()

	if exists {
		// 发送取消信号
		task.cancel()
	}
}

// StopAll 停止所有服务的跟踪
func (t *Tracer) StopAll() {
	t.mu.RLock()
	tasksCopy := make([]*taskControl, 0, len(t.tasks))
	for _, task := range t.tasks {
		tasksCopy = append(tasksCopy, task)
	}
	t.mu.RUnlock()

	// 取消所有任务
	for _, task := range tasksCopy {
		task.cancel()
	}
}
