package dto

// ResourceRequirements 定义资源需求
type ResourceRequirements struct {
	AcceleratorType  string `json:"acceleratorType"`  // 显卡类型
	AcceleratorCount int    `json:"acceleratorCount"` // 显卡数量
}

// DeploySpec 部署期望结构体
type DeploySpec struct {
	ServiceId            string                `json:"serviceId"`
	ModelName            string                `json:"modelName"`
	ModelFileDir         string                `json:"modelFileDir"`
	ResourceRequirements *ResourceRequirements `json:"resourceRequirements"`
	ReplicaCount         int                   `json:"replicaCount"`
	ContextLength        int                   `json:"contextLength"`
	Env                  []Env                 `json:"env"`
	GoalSetName          string                `json:"goalSetName"`
}

type Env struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
