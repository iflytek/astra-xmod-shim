// internal/workqueue/queue.go

package workqueue

import (
	"time"

	"k8s.io/client-go/util/workqueue"
)

// Queue 是一个类型安全的字符串工作队列，封装了 client-go 的 TypedRateLimitingInterface。
type Queue struct {
	wq workqueue.TypedRateLimitingInterface[string]
}

// New 创建一个带默认指数退避限流器的命名工作队列。
// 默认配置：初始重试延迟 5ms，最大延迟 1000 秒。
func New() *Queue {
	rateLimiter := workqueue.NewTypedItemExponentialFailureRateLimiter[string](
		5*time.Millisecond,
		1000*time.Second,
	)
	q := workqueue.NewTypedRateLimitingQueueWithConfig(
		rateLimiter,
		workqueue.TypedRateLimitingQueueConfig[string]{},
	)
	return &Queue{wq: q}
}

// Add 将一个键加入队列。
func (q *Queue) Add(key string) {
	q.wq.Add(key)
}

// AddAfter 在指定延迟后将键加入队列。
func (q *Queue) AddAfter(key string, duration time.Duration) {
	q.wq.AddAfter(key, duration)
}

// Get 从队列中取出一个键。
// 返回值：
//   - key: 队列中的键
//   - done: 调用此函数表示该项已处理完成（必须调用！）
//
// 如果队列已关闭，会 panic（通常只在程序退出时发生）。
func (q *Queue) Get() (key string, done func()) {
	item, shutdown := q.wq.Get()
	if shutdown {
		panic("workqueue: Get() called after queue was shut down")
	}
	return item, func() { q.wq.Done(item) }
}

// Len 返回当前队列中待处理项的数量。
func (q *Queue) Len() int {
	return q.wq.Len()
}

// ShutDown 关闭队列，禁止后续 Add 操作，并让 Get 在队列空时立即返回。
// 关闭后再次调用 Get 会返回 (zero, true)。
func (q *Queue) ShutDown() {
	q.wq.ShutDown()
}

// ShutDownWithDrain 关闭队列，但允许已入队的项继续被处理（推荐用于优雅关闭）。
func (q *Queue) ShutDownWithDrain() {
	q.wq.ShutDownWithDrain()
}

// Forget 表示该项处理成功，清除其重试计数。
// 下次再 Add 相同 key 时，将从初始延迟开始重试。
func (q *Queue) Forget(key string) {
	q.wq.Forget(key)
}

// NumRequeues 返回该键已被重试的次数。
func (q *Queue) NumRequeues(key string) int {
	return q.wq.NumRequeues(key)
}
