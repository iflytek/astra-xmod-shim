package bootstrap

import "testing"

func TestMain(m *testing.M) {
	m.Run()
}

// 测试正常配置加载流程
func TestBootStrap_Success(t *testing.T) {
	Init("/Users/haoxuanli/Documents/GitHub/iflytek/astron-xmod-shim/conf.yaml")
	//// 验证结果
	//assert.NoError(t, err)
	//assert.NotNil(t, cfg)
	//assert.Equal(t, "8080", cfg.Server.Port)
	//assert.Equal(t, "info", cfg.Log.Level)
	//assert.Equal(t, "./logs", cfg.Log.Path)
}
