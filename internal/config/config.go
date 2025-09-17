package config

import (
	"fmt"
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
