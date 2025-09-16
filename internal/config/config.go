package config

import (
	"fmt"
	confSpec "modserv-shim/internal/dto/config"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var (
	globalConfig *confSpec.GlobalConfig // 修正：使用model.Config（与返回类型一致）
	once         sync.Once
	initErr      error
	configPath   string
	mu           sync.RWMutex // 新增：用于热更新时的线程安全
)

// SetConfigPath 提前设置配置文件路径
func SetConfigPath(path string) {
	configPath = path
}

// Get 懒加载获取配置实例（线程安全）
func Get() (*confSpec.GlobalConfig, error) {
	// 先读锁检查，避免每次加写锁
	mu.RLock()
	if globalConfig != nil || initErr != nil {
		mu.RUnlock()
		return globalConfig, initErr
	}
	mu.RUnlock()

	once.Do(func() {
		path := configPath
		v := viper.New()
		v.SetConfigFile(path)
		v.SetConfigType("yaml")

		// 读取配置文件
		if err := v.ReadInConfig(); err != nil {
			initErr = fmt.Errorf("读取配置文件失败: %w", err)
			return
		}

		// 解析配置到结构体
		globalConfig = &confSpec.GlobalConfig{}
		if err := v.Unmarshal(globalConfig); err != nil {
			initErr = fmt.Errorf("解析配置失败: %w", err)
			globalConfig = nil
			return
		}

		// 启用配置热更新（加锁保证线程安全）
		v.WatchConfig()
		v.OnConfigChange(func(e fsnotify.Event) {
			fmt.Printf("配置文件已更新: %s\n", e.Name)
			newConfig := &confSpec.GlobalConfig{}
			if err := v.Unmarshal(newConfig); err != nil {
				fmt.Printf("配置热更新失败: %v\n", err)
				return
			}
			// 替换全局配置时加写锁
			mu.Lock()
			globalConfig = newConfig
			mu.Unlock()
		})
	})

	return globalConfig, initErr
}
