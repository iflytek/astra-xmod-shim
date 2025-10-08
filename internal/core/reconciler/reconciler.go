package reconciler

import (
	"astron-xmod-shim/internal/core/goal"
	"astron-xmod-shim/internal/core/spec"
	"astron-xmod-shim/internal/core/workqueue"
	dto "astron-xmod-shim/internal/dto/deploy"
	"context"
	"sync"
)

type Reconciler struct {
	queue     *workqueue.Queue
	specStore spec.Store // 见下文说明
	workers   int
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

// NewReconciler 创建一个可运行的 reconciler
func NewReconciler(store spec.Store, workers int) *Reconciler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Reconciler{
		queue:     workqueue.New(),
		specStore: store,
		workers:   workers,
		ctx:       ctx,
		cancel:    cancel,
	}
}

func (r *Reconciler) reconcile(spec *dto.DeploySpec) error {

	// 组装 ctx
	goalSetCtx := &goal.Context{
		Data: make(map[string]any),
	}

	// 获取 goals
	goalSet := goal.Registry[spec.GoalSetName]

	// run goals
	for _, singleGoal := range goalSet.Goals {
		if !singleGoal.IsAchieved(goalSetCtx) {
			err := singleGoal.Ensure(goalSetCtx)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Start 启动消费者协程
func (r *Reconciler) Start() {
	for i := 0; i < r.workers; i++ {
		r.wg.Add(1)
		go r.runWorker()
	}
}

func (r *Reconciler) runWorker() {
	defer r.wg.Done()
	for {
		select {
		case <-r.ctx.Done():
			return // 优雅退出
		default:
		}

		key, done := r.queue.Get()
		deploySpec := r.specStore.Get(key)
		err := r.reconcile(deploySpec)
		if err != nil {
			// 注意：workqueue 会自动重试（因为没调用 Forget）
		} else {
			r.queue.Forget(key) // 清除重试计数
		}

		done() // 告诉队列这项已完成
	}
}
