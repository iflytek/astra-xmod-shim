package shimlet_test

import (
	"astron-xmod-shim/internal/core/shimlet"
	_ "astron-xmod-shim/internal/core/shimlet/shimlets"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	exitCode := m.Run()
	os.Exit(exitCode)
}

// 测试正常配置加载流程
func TestGetReg(t *testing.T) {
	a, _ := shimlet.Registry.GetSingleton("k8s")
	if a != nil {

	}
}
