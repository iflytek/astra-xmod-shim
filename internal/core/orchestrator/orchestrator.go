package orchestrator

import (
	"astron-xmod-shim/internal/config"
	"astron-xmod-shim/internal/core/goal"
	_ "astron-xmod-shim/internal/core/goal/goalset"
	"astron-xmod-shim/internal/core/shimlet"
	"astron-xmod-shim/internal/core/spec"
	"astron-xmod-shim/internal/core/typereg"
	"astron-xmod-shim/internal/core/workqueue"
	dto "astron-xmod-shim/internal/dto/deploy"
	"astron-xmod-shim/pkg/log"
	"fmt"
)

type Orchestrator struct {
	shimReg    *typereg.TypeReg[shimlet.Shimlet]
	goalSetReg map[string]*goal.GoalSet
	specStore  spec.Store
	queue      *workqueue.Queue
}

func NewOrchestrator(
	shimReg *typereg.TypeReg[shimlet.Shimlet],
	pipeReg map[string]*goal.GoalSet,
	queue *workqueue.Queue,
	specStore spec.Store,
) *Orchestrator {
	return &Orchestrator{
		queue:      queue,
		shimReg:    shimReg,
		goalSetReg: pipeReg,
		specStore:  specStore,
	}
}

var GlobalOrchestrator *Orchestrator

func (o *Orchestrator) Provision(spec *dto.RequirementSpec) error {

	// 覆盖掉 nvidia.com/gpu 的 limit
	spec.ResourceRequirements.AcceleratorType = "nvidia.com/gpu"

	// goalset 已在api handler 层 确定
	// shimlet 已在启动时配置全局确定

	// RequirementSpec 持久化 部署期望
	spec.ReplicaCount = 1
	spec.ShimletName = config.Get().CurrentShimlet
	// 如果这里是更新, 则需要 对应goalset reconcile 检测到 不一致 并调用ensure 闭环
	o.specStore.Set(spec.ServiceId, spec)

	// 投递到队列
	o.queue.Add(spec.ServiceId)

	return nil
}

// DeleteService 删除指定的模型服务
func (o *Orchestrator) DeleteService(serviceID string) error {
	// 获取当前使用的shimlet
	currentShimletId := config.Get().CurrentShimlet
	runtimeShimlet, err := o.shimReg.GetSingleton(currentShimletId)
	if err != nil {
		log.Error("get runtime shimlet error", err)
		return err
	}

	// 调用shimlet的Delete方法删除资源
	if err := runtimeShimlet.Delete(serviceID); err != nil {
		log.Error("delete service failed", err)
		return err
	}

	go log.Info("service deleted successfully", "serviceID", serviceID)
	return nil
}

// GetServiceStatus 获取指定服务的状态信息
func (o *Orchestrator) GetServiceStatus(serviceID string) (*dto.RuntimeStatus, error) {
	if serviceID == "" {
		return nil, fmt.Errorf("serviceID is required")
	}

	currentShimletId := config.Get().CurrentShimlet
	runtimeShimlet, err := o.shimReg.GetSingleton(currentShimletId)
	if err != nil {
		return nil, err
	}
	status, err := runtimeShimlet.Status(serviceID)
	if err != nil {
		return nil, err
	}
	if status.EndPoint != "" {
		status.EndPoint += "/v1/chat/completions"
	}
	return status, nil
}
