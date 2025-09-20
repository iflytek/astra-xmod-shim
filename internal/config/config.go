package config

import (
	"fmt"
	"os"
	"sync"

	confSpec "modserv-shim/internal/dto/config"

	"github.com/spf13/viper"
)

var (
	globalConfig *confSpec.GlobalConfig
	once         sync.Once
	initErr      error
	configPath   string
)

// SetConfigPath 提前设置配置文件路径
func SetConfigPath(path string) {
	configPath = path
}

// Get 懒加载获取配置实例（线程安全）
func Get() *confSpec.GlobalConfig {
	once.Do(func() {
		if configPath == "" {
			initErr = fmt.Errorf("config path not set")
			return
		}

		v := viper.New()
		v.SetConfigFile(configPath)
		v.SetConfigType("yaml")

		if err := v.ReadInConfig(); err != nil {
			initErr = fmt.Errorf("读取配置文件失败: %w", err)
			return
		}

		globalConfig = &confSpec.GlobalConfig{}
		if err := v.Unmarshal(globalConfig); err != nil {
			initErr = fmt.Errorf("解析配置失败: %w", err)
			globalConfig = nil
			return
		}
	})

	// 加载失败也返回 nil，符合预期
	if initErr != nil {
		// 可选：打印错误日志，或通过其他方式暴露 initErr
		// log.Printf("配置加载失败: %v", initErr)
		return nil
	}

	return globalConfig
}

// GetConfFromFileDir 从指定文件路径加载特定类型的配置
// 支持加载YAML格式的配置文件
// configPath: 配置文件的完整路径
// 返回加载后的配置实例
func GetConfFromFileDir[T any](configPath string) (*T, error) {
	// 创建T类型的新实例并获取其指针
	conf := new(T)
	// 检查文件是否存在
	stat, err := os.Stat(configPath)
	if err != nil {
		return nil, fmt.Errorf("配置文件不存在: %w", err)
	}
	if stat.IsDir() {
		return nil, fmt.Errorf("路径不是文件: %s", configPath)
	}

	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 将配置解析到新创建的结构体指针中
	if err := v.Unmarshal(conf); err != nil {
		return nil, fmt.Errorf("解析配置到结构体失败: %w", err)
	}

	return conf, nil
}
