package conf

// Config 应用主配置结构体，显式添加mapstructure标签
type Config struct {
	K8s        K8sConfig  `yaml:"k8s" mapstructure:"k8s"`
	HTTPServer HTTPServer `yaml:"http-server" mapstructure:"http-server"`
	Log        LogConfig  `yaml:"log" mapstructure:"log"`
}

// K8sConfig Kubernetes客户端配置
type K8sConfig struct {
	Kubeconfig string  `yaml:"kube-config" mapstructure:"kube-config"`
	Context    string  `yaml:"context" mapstructure:"context"`
	QPS        float32 `yaml:"qps" mapstructure:"qps"`
	Burst      int     `yaml:"burst" mapstructure:"burst"`
	Timeout    int64   `yaml:"timeout" mapstructure:"timeout"`
}

// HTTPServer HTTP服务器配置
type HTTPServer struct {
	Port string `yaml:"port" mapstructure:"port"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level         string `yaml:"level" mapstructure:"level"`
	OutputPath    string `yaml:"output-path" mapstructure:"output-path"`
	MaxSize       int    `yaml:"max-size" mapstructure:"max-size"`
	MaxAge        int    `yaml:"max-age" mapstructure:"max-age"`
	Compress      bool   `yaml:"compress" mapstructure:"compress"`
	ShowLine      bool   `yaml:"show-line" mapstructure:"show-line"`
	EnableConsole bool   `yaml:"enable-console" mapstructure:"enable-console"`
}
