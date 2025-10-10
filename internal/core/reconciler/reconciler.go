package reconciler

import (
	"astron-xmod-shim/internal/core/goal"
	"astron-xmod-shim/internal/core/shimlet"
	"astron-xmod-shim/internal/core/spec"
	"astron-xmod-shim/internal/core/workqueue"
	dto "astron-xmod-shim/internal/dto/deploy"
	"context"
	"errors"
	"sync"
	"time"
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
func NewReconciler(store spec.Store, workers int, queue *workqueue.Queue) *Reconciler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Reconciler{
		queue:     queue,
		specStore: store,
		workers:   workers,
		ctx:       ctx,
		cancel:    cancel,
	}
}

func (r *Reconciler) reconcile(spec *dto.RequirementSpec) error {

	// 组装 ctx
	infraShim, err := shimlet.Registry.GetSingleton(spec.ShimletName)
	if err != nil {
		return err
	}

	goalSetCtx := &goal.Context{
		Data:       make(map[string]any),
		DeploySpec: spec,
		Shimlet:    infraShim,
	}

	// 获取 goals
	goalSet := goal.Registry[spec.GoalSetName]

	// 这里遍历全部goal 严格按先后顺序遍历
	for _, singleGoal := range goalSet.Goals {
		// 如果有goal 没有达成 则调用 ensure
		if !singleGoal.IsAchieved(goalSetCtx) {
			err := singleGoal.Ensure(goalSetCtx)
			if err != nil {
				return err
			}
		}
		if !singleGoal.IsAchieved(goalSetCtx) {
			return errors.New("some goal not yet serviceId:" + goalSetCtx.DeploySpec.ServiceId)
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
		func() {
			key, done := r.queue.Get()

			defer done()

			select {
			case <-r.ctx.Done():
				return // 优雅退出
			default:
			}
			deploySpec := r.specStore.Get(key)
			err := r.reconcile(deploySpec)
			if err != nil {
				r.queue.Forget(key) // 清除重试计数
				r.queue.AddAfter(key, time.Second*10)
				// 注意：workqueue 会自动重试（因为没调用 Forget）
			} else {
				r.queue.Forget(key) // 清除重试计数
				r.queue.AddAfter(key, time.Second*300)
			}

		}()

	}
}
