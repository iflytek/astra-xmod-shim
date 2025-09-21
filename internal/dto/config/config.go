package dto

// GlobalConfig 应用主配置结构体，显式添加mapstructure标签
type GlobalConfig struct {
	K8s            K8sConfig                `yaml:"k8s" mapstructure:"k8s"`
	Server         Server                   `yaml:"server" mapstructure:"server"`
	Log            LogConfig                `yaml:"log" mapstructure:"log"`
	CurrentShimlet string                   `yaml:"current-shimlet" mapstructure:"current-shimlet"` // 👈 新增
	Shimlets       map[string]ShimletConfig `yaml:"shimlets" mapstructure:"shimlets"`
	ModelManage    ModelManageConfig        `yaml:"model-manage" mapstructure:"model-manage"`
}

// K8sConfig Kubernetes客户端配置
type K8sConfig struct {
	Kubeconfig string  `yaml:"kube-config" mapstructure:"kube-config"`
	Context    string  `yaml:"context" mapstructure:"context"`
	QPS        float32 `yaml:"qps" mapstructure:"qps"`
	Burst      int     `yaml:"burst" mapstructure:"burst"`
	Timeout    int64   `yaml:"timeout" mapstructure:"timeout"`
}

// Server HTTP服务器配置
type Server struct {
	Port string `yaml:"port" mapstructure:"port"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level         string `yaml:"level" mapstructure:"level"`
	Path          string `yaml:"path" mapstructure:"path"`
	MaxSize       int    `yaml:"max-size" mapstructure:"max-size"`
	MaxAge        int    `yaml:"max-age" mapstructure:"max-age"`
	Compress      bool   `yaml:"compress" mapstructure:"compress"`
	ShowLine      bool   `yaml:"show-line" mapstructure:"show-line"`
	EnableConsole bool   `yaml:"enable-console" mapstructure:"enable-console"`
}

// ShimletConfig 插件配置（动态）
type ShimletConfig struct {
	ConfigPath string `yaml:"config-path" mapstructure:"config-path"`
}

// ModelManageConfig 模型管理配置
type ModelManageConfig struct {
	ModelRoot string `yaml:"model-root" mapstructure:"model-route"`
}
