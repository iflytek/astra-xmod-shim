package tracer

import (
	"context"
	"modserv-shim/internal/core/shimlet"
	"modserv-shim/pkg/log"
	"sync"
	"time"
)

// Tracer 轻量级状态跟踪器
// 直接维护协程上下文的map，无需额外的tracker层

type Tracer struct {
	// 跟踪任务映射：serviceID -> 任务控制信息
	tasks map[string]*taskControl
	mu    sync.RWMutex // 保护tasks的读写锁
}

// taskControl 封装单个跟踪任务的控制信息
// 这是一个内部结构体，无需暴露给外部

type taskControl struct {
	serviceID string             // 服务ID
	shimlet   shimlet.Shimlet    // shimlet实例
	cancel    context.CancelFunc // 协程取消函数
}

// NewTracer 创建一个新的轻量级Tracer实例

func NewTracer() *Tracer {
	return &Tracer{
		tasks: make(map[string]*taskControl),
	}
}

// Trace 方法直接启动一个定时协程任务
// serviceID: 服务实例ID
// shim: 要调用的shimlet实例
// interval: 状态检查间隔时间

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

// 跟踪服务的协程函数，直接在Tracer中实现
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

// 执行服务状态检查
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
