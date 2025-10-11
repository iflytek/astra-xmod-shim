package goalset

import (
	"astron-xmod-shim/internal/core/goal"
	dto "astron-xmod-shim/internal/dto/deploy"
	"time"
)

func init() {
	NewLLMDeleteGoalSet()
}

var deployDeleted = goal.Goal{Name: "deployFinish",
	IsAchieved: func(ctx *goal.Context) bool {
		status, err := ctx.Shimlet.Status(ctx.DeploySpec.ServiceId)
		if err != nil {
			return false
		}
		return status.Status == dto.PhaseUnknown || status.Status == dto.PhaseTerminated
	},
	Ensure: func(ctx *goal.Context) error {
		err := ctx.Shimlet.Delete(ctx.DeploySpec.ServiceId)
		if err != nil {
			return err
		}
		return nil
	}}

// NewLLMDeleteGoalSet 创建一个用于下线 LLM 模型的 GoalSet
func NewLLMDeleteGoalSet() {
	goal.NewGoalSetBuilder("opensource-llm-delete").
		AddGoal(deployDeleted).
		WithMaxRetries(10).           // 失败最多重试 3 次
		WithTimeout(5 * time.Minute). // 整体超时 2 分钟
		BuildAndRegister()
}
