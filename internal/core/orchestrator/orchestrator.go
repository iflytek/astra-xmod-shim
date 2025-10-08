package orchestrator

import (
	"fmt"
	"modserv-shim/internal/config"
	"modserv-shim/internal/core/goal"
	_ "modserv-shim/internal/core/goal/goalset"
	"modserv-shim/internal/core/shimlet"
	"modserv-shim/internal/core/state"
	"modserv-shim/internal/core/typereg"
	"modserv-shim/internal/core/workqueue"
	dto "modserv-shim/internal/dto/deploy"
	"modserv-shim/pkg/log"
)

type Orchestrator struct {
	shimReg    *typereg.TypeReg[shimlet.Shimlet]
	goalSetReg map[string]*goal.GoalSet
	stateMgr   *state.Manager
	queue      *workqueue.Queue
}

func NewOrchestrator(
	shimReg *typereg.TypeReg[shimlet.Shimlet],
	pipeReg map[string]*goal.GoalSet,
	queue *workqueue.Queue,
	stateMgr *state.Manager,
) *Orchestrator {
	return &Orchestrator{
		queue:      queue,
		shimReg:    shimReg,
		goalSetReg: pipeReg,
		stateMgr:   stateMgr,
	}
}

var GlobalOrchestrator *Orchestrator

func (o *Orchestrator) Provision(spec *dto.DeploySpec) error {

	// 覆盖掉 nvidia.com/gpu 的 limit
	spec.ResourceRequirements.AcceleratorType = "nvidia.com/gpu"

	// DeploySpec 需要持久化 这是用户的部署期望 TODO: 持久化

	// TODO 组装用户部署期望spec
	infraShim, err := o.shimReg.GetSingleton(config.Get().CurrentShimlet)
	if err != nil {
		return err
	}
	spec.Shimlet = infraShim
	spec.GoalSet, _ = o.goalSetReg[getGoalSetName()] // 目标集

	o.stateMgr.Set(spec.ServiceId, spec)

	// TODO 投递到队列
	o.queue.Add(spec.ServiceId)

	_ = &goal.Context{
		ResourceId: spec.ServiceId,
		Queue:      o.queue,
		Data:       make(map[string]any),
	}
	// 4. 执行pipeline

	return nil
}

// TODO: [临时] 后续应根据 spec.Type 或 metadata 动态选择 pipeline
// 示例：spec.Type == "llm" → "ai-pipeline", spec.Type == "web" → "web-pipeline"
func getGoalSetName() string {
	return "opensource_llm"
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

	o.stateMgr.GetStatus(serviceID)

	// 如果没找到，可以返回“不存在”，或 fallback 到远程查询（可选）

	return nil, nil
}
