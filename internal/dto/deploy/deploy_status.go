package dto

// DeployStatus 部署状态
type DeployStatus struct {
	DeploySpec DeploySpec  `json:"modelFile"`
	Status     DeployPhase `json:"contextLength"`
}
