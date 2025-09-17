package main

import (
	"log"
	"modserv-shim/internal/bootstrap"
	_ "modserv-shim/internal/shimreg/shimlets" // 显式导入插件依赖包
	"os"

	"github.com/spf13/cobra"
)

var configPath string

func main() {
	rootCmd := &cobra.Command{
		Use:   "modserv-shim",
		Short: "model serve shim",
		RunE:  runMw,
	}

	// 注册配置文件参数
	rootCmd.Flags().StringVarP(
		&configPath,
		"config", "c",
		"conf.yaml",
		"配置文件路径",
	)

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("启动失败: %v", err)
	}
}

// runMw 启动中间件和注册退出钩子
func runMw(cmd *cobra.Command, args []string) error {
	// 1. 验证配置文件
	if err := validateConfigFile(configPath); err != nil {
		return err
	}
	log.Printf("use cfg from: %s", configPath)

	// 2. bootstrap
	if err := bootstrap.Init(configPath); err != nil {
		return err
	}

	// 3. 阻塞等待退出信号
	waitForShutdownSignal()
	return nil
}

// validateConfigFile 验证配置文件是否存在
func validateConfigFile(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}
	return nil
}

// waitForShutdownSignal 阻塞等待退出信号
func waitForShutdownSignal() {
	// 等待析构 waitGroup 完毕回调
	bootstrap.WaitForShutDown()
	log.Println("receive shutdown signal waiting for resource release...")
}
