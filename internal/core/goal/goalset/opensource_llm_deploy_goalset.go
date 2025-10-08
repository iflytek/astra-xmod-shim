package goalset

import (
	"modserv-shim/internal/config"
	"modserv-shim/internal/core/goal"
	"modserv-shim/pkg/log"
	"path/filepath"
	"time"
)

// 构造 mapModelNameToPath Goal
var modelPathReady = goal.Goal{
	Name: "map-model-path",
	IsAchieved: func(ctx *goal.Context) bool {
		// 如果 ModelFileDir 已设置，说明已经执行过

		return false
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
		ctx.Data["model-path"] = modelDir
		return nil
	},
}

var deployFinish = goal.Goal{Name: "deployFinish",
	IsAchieved: func(ctx *goal.Context) bool {
		if ctx.DeploySpec.ServiceId != "" {
			return true
		}
		return false
	},
	Ensure: func(ctx *goal.Context) error {
		err := ctx.Shimlet.Apply(ctx.DeploySpec)
		if err != nil {
			return err
		}
		return nil
	}}

// NewLLMDeployGoalSet 创建一个用于部署 LLM 模型的 GoalSet
func NewLLMDeployGoalSet() *goal.GoalSet {
	return goal.NewGoalSetBuilder("llm-deploy").
		AddGoal(modelPathReady).
		AddGoal(deployFinish).
		WithMaxRetries(10).           // 失败最多重试 10 次
		WithTimeout(5 * time.Minute). // 整体超时 5 分钟
		Build()
}
