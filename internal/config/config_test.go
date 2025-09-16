package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	SetConfigPath("666")
	exitCode := m.Run()
	os.Exit(exitCode)

}

// 测试正常配置加载流程
func TestGetConfig_Success(t *testing.T) {
	// 创建临时测试配置文件
	content := `
server:
  port: 8080
  timeout: 30s
log:
  level: info
  path: "./logs"
`
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name()) // 清理临时文件

	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())

	// 设置配置文件路径并加载
	SetConfigPath(tmpFile.Name())
	cfg, err := Get()

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "8080", cfg.Server.Port)
	assert.Equal(t, "info", cfg.Log.Level)
	assert.Equal(t, "./logs", cfg.Log.Path)
}
