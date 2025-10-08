package goalset

import (
	"modserv-shim/internal/core/goal"
	"time"
)

var deployDeleted = goal.Goal{Name: "deployFinish",
	IsAchieved: func(ctx goal.Context) bool {
		status, err := ctx.Shimlet.Status(ctx.ResourceId)
		if err != nil {
			return false
		}
		return status.EndPoint == ""
	},
	Ensure: func(ctx goal.Context) error {
		err := ctx.Shimlet.Delete(ctx.ResourceId)
		if err != nil {
			return err
		}
		return nil
	}}

// NewLLMDeleteGoalSet 创建一个用于下线 LLM 模型的 GoalSet
func NewLLMDeleteGoalSet() *goal.GoalSet {
	return goal.NewGoalSetBuilder("llm-deploy").
		AddGoal(modelPathReady).
		AddGoal(deployFinish).
		WithMaxRetries(10).           // 失败最多重试 3 次
		WithTimeout(5 * time.Minute). // 整体超时 2 分钟
		Build()
}
