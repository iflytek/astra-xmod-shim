package model

// ResourceRequirements 定义资源需求
type ResourceRequirements struct {
	AcceleratorType  string `json:"acceleratorType"`  // 显卡类型
	AcceleratorCount int    `json:"acceleratorCount"` // 显卡数量
}

// DeploySpec 模型部署请求结构体
type DeploySpec struct {
	ModelFile            string               `json:"modelFile"`
	ResourceRequirements ResourceRequirements `json:"resourceRequirements"`
	ReplicaCount         int                  `json:"replicaCount"`
	ContextLength        int                  `json:"contextLength"`
	Env                  []Env                `json:"env"`
}
type Env struct {
	Key string
	Val string
}

type EnvVar struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
