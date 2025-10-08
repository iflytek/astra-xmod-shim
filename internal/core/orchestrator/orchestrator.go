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

func (o *Orchestrator) Provision(spec *dto.DeploySpec) error {

	// 覆盖掉 nvidia.com/gpu 的 limit
	spec.ResourceRequirements.AcceleratorType = "nvidia.com/gpu"

	// DeploySpec 需要持久化 这是用户的部署期望 TODO: 持久化

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

	//o.specStore.GetStatus(serviceID)

	// 如果没找到，可以返回“不存在”，或 fallback 到远程查询（可选）

	return nil, nil
}
