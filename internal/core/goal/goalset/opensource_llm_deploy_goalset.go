package goalset

import (
	"astron-xmod-shim/internal/config"
	"astron-xmod-shim/internal/core/goal"
	dto "astron-xmod-shim/internal/dto/deploy"
	"astron-xmod-shim/pkg/log"
	"path/filepath"
	"reflect"
	"time"
)

func init() {
	NewLLMDeployGoalSet()
}

// areResourceRequirementsEqual 比较两个 ResourceRequirements 是否相等
// 处理 nil 和零值的情况
func areResourceRequirementsEqual(expected, actual *dto.ResourceRequirements) bool {
	// 如果都为 nil，认为相等
	if expected == nil && actual == nil {
		return true
	}

	// 如果其中一个为 nil，另一个不为 nil，但所有字段都是零值，也认为相等
	if expected == nil && actual != nil {
		return actual.AcceleratorType == "" && actual.AcceleratorCount == 0
	}

	if expected != nil && actual == nil {
		return expected.AcceleratorCount == 0
	}

	// 都不为 nil，使用 reflect.DeepEqual 比较
	return reflect.DeepEqual(expected, actual)
}

// 构造 mapModelNameToPath Goal
var modelPathReady = goal.Goal{
	Name: "map-model-path",
	IsAchieved: func(ctx *goal.Context) bool {
		// 如果 ModelFileDir 已设置，说明已经执行过

		return ctx.DeploySpec.ModelFileDir != ""
	},
	Ensure: func(ctx *goal.Context) error {
		modelRoot := config.Get().ModelManage.ModelRoot
		if modelRoot == "" {
			modelRoot = "/models"
			log.Info("Using default model root directory: %s", modelRoot)
		}

		// Construct full model path
		modelDir := filepath.Join(modelRoot, ctx.DeploySpec.ModelName)
		log.Info("Mapping model name to path: %s -> %s", ctx.DeploySpec.ModelName, modelDir)

		// Set mapped path to DeploySpec
		ctx.DeploySpec.ModelFileDir = modelDir
		return nil
	},
}

var specConsistencyCheck = goal.Goal{
	Name: "spec-consistency-check",
	IsAchieved: func(ctx *goal.Context) bool {
		// 如果还没有serviceId，说明还没部署过，不需要检查一致性
		if ctx.DeploySpec.ServiceId == "" {
			return true
		}

		// 获取当前运行时状态
		status, err := ctx.Shimlet.Status(ctx.DeploySpec.ServiceId)
		if err != nil {
			log.Warn("Failed to get status for service %s: %v", ctx.DeploySpec.ServiceId, err)
			return false
		}

		// 比较期望的spec和实际的spec是否一致
		// 这里我们主要比较关键字段：ModelName, ModelFileDir, ResourceRequirements, ReplicaCount
		expectedSpec := ctx.DeploySpec
		actualSpec := status.DeploySpec

		// 比较关键字段
		if expectedSpec.ModelName != actualSpec.ModelName ||
			expectedSpec.ReplicaCount != actualSpec.ReplicaCount ||
			!areResourceRequirementsEqual(expectedSpec.ResourceRequirements, actualSpec.ResourceRequirements) {
			log.Info("Spec inconsistency detected for service %s", ctx.DeploySpec.ServiceId)
			return false
		}

		return true
	},
	Ensure: func(ctx *goal.Context) error {
		log.Info("Re-applying spec for service %s due to inconsistency", ctx.DeploySpec.ServiceId)
		// 如果spec不一致，重新应用当前的spec
		return ctx.Shimlet.Apply(ctx.DeploySpec)
	},
}

var deployFinished = goal.Goal{Name: "deployFinish",
	IsAchieved: func(ctx *goal.Context) bool {
		status, err := ctx.Shimlet.Status(ctx.DeploySpec.ServiceId)
		if err != nil {
			return false
		}
		return status.Status != dto.PhaseUnknown
	},
	Ensure: func(ctx *goal.Context) error {
		err := ctx.Shimlet.Apply(ctx.DeploySpec)
		if err != nil {
			return err
		}
		return nil
	}}

var serviceExposed = goal.
	Goal{Name: "exposeService",
	IsAchieved: func(ctx *goal.Context) bool {
		return ctx.DeploySpec.ServiceId != ""
	},
	Ensure: func(ctx *goal.Context) error {
		status, err := ctx.Shimlet.Status(ctx.DeploySpec.ServiceId)
		if err != nil || status.EndPoint == "" {
			return err
		}
		return nil
	}}

// NewLLMDeployGoalSet 创建一个用于部署 LLM 模型的 GoalSet
func NewLLMDeployGoalSet() {
	goal.NewGoalSetBuilder("opensource-llm-deploy").
		AddGoal(modelPathReady).
		AddGoal(deployFinished).
		AddGoal(specConsistencyCheck). // 添加spec一致性检查Goal
		AddGoal(serviceExposed).
		WithMaxRetries(10).           // 失败最多重试 10 次
		WithTimeout(5 * time.Minute). // 整体超时 5 分钟
		BuildAndRegister()
}
