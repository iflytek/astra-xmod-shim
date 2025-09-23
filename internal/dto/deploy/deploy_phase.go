package dto

// DeployPhase Deployment status enumeration (recommended to place in a separate status.go file)
// DeployPhase defines all possible states (phases) of deployment
type DeployPhase string

const (
	PhaseUnknown     DeployPhase = "unknown" // ✅ 新增
	PhasePending     DeployPhase = "pending"
	PhaseCreating    DeployPhase = "creating"
	PhaseRunning     DeployPhase = "running"
	PhaseFailed      DeployPhase = "failed"
	PhaseTerminating DeployPhase = "terminating"
	PhaseTerminated  DeployPhase = "terminated"
)
