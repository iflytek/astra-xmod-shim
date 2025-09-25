package orchestrator

import (
	"fmt"
	"modserv-shim/internal/config"
	"modserv-shim/internal/core/eventbus"
	"modserv-shim/internal/core/pipeline"
	_ "modserv-shim/internal/core/pipeline/pipelines"
	"modserv-shim/internal/core/shimlet"
	"modserv-shim/internal/core/statemanager"
	"modserv-shim/internal/core/tracer"
	"modserv-shim/internal/core/typereg"
	dto "modserv-shim/internal/dto/deploy"
	eventbus2 "modserv-shim/internal/dto/eventbus"
	"modserv-shim/pkg/log"
)

type Orchestrator struct {
	shimReg      *typereg.TypeReg[shimlet.Shimlet]
	pipeReg      map[string]*pipeline.Pipeline
	eventBus     eventbus.EventBus
	stateManager *statemanager.StateManager
	tracer       *tracer.Tracer
}

func NewOrchestrator(
	shimReg *typereg.TypeReg[shimlet.Shimlet],
	pipeReg map[string]*pipeline.Pipeline,
	eventBus eventbus.EventBus,
	tracer *tracer.Tracer,
	stateManager *statemanager.StateManager,
) *Orchestrator {
	return &Orchestrator{
		shimReg:      shimReg,
		pipeReg:      pipeReg,
		eventBus:     eventBus,
		tracer:       tracer,
		stateManager: stateManager,
	}
}

var GlobalOrchestrator *Orchestrator

func (o *Orchestrator) Provision(spec *dto.DeploySpec) error {

	// 兜底事件：开始部署
	o.eventBus.Publish("service.status", &eventbus2.ServiceEvent{
		ServiceID: spec.ServiceId,
		To:        dto.PhaseCreating,
	})

	// 覆盖掉 nvidia.com/gpu 的 limit
	spec.ResourceRequirements.AcceleratorType = "nvidia.com/gpu"

	runtimePipe := o.pipeReg[getPipelineName()]
	if runtimePipe == nil {
		return fmt.Errorf("pipeline %s not found", getPipelineName())
	}

	currentShimletId := config.Get().CurrentShimlet
	runtimeShimlet, err := o.shimReg.GetSingleton(currentShimletId)
	if err != nil {
		log.Error("get runtime shimlet error", err)
		return err
	}
	pipeCtx := &pipeline.Context{
		Shimlet:    runtimeShimlet,
		DeploySpec: spec,
		EventBus:   o.eventBus,
		ResourceId: spec.ServiceId,
		Tracer:     o.tracer,
		Data:       make(map[string]any),
	}
	// 4. 执行pipeline
	if err := runtimePipe.Execute(pipeCtx); err != nil {
		// 兜底事件：失败
		o.eventBus.Publish("service.status", &eventbus2.ServiceEvent{
			ServiceID: spec.ServiceId,
			To:        dto.PhaseFailed,
		})
		return err
	}

	return nil
}

// TODO: [临时] 后续应根据 spec.Type 或 metadata 动态选择 pipeline
// 示例：spec.Type == "llm" → "ai-pipeline", spec.Type == "web" → "web-pipeline"
func getPipelineName() string {
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

	runtimeStat := o.stateManager.GetStatus(serviceID)

	// 如果没找到，可以返回“不存在”，或 fallback 到远程查询（可选）

	return runtimeStat, nil
}
