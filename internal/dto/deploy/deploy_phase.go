package dto

// DeployPhase 部署状态枚举（建议放在单独的 status.go 中）
// DeployPhase 定义部署的所有可能状态（阶段）
type DeployPhase string

const (
	PhasePending     DeployPhase = "PENDING"     // 初始化中，等待调度
	PhaseCreating    DeployPhase = "CREATING"    // 资源创建中
	PhaseRunning     DeployPhase = "RUNNING"     // 部署成功，正常运行
	PhaseUpdating    DeployPhase = "UPDATING"    // 正在更新配置
	PhaseFailed      DeployPhase = "FAILED"      // 部署失败
	PhaseTerminating DeployPhase = "TERMINATING" // 正在终止/删除
	PhaseTerminated  DeployPhase = "TERMINATED"  // 已终止/删除
)
