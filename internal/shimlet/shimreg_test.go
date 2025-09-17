package shimlet_test

import (
	"modserv-shim/internal/shimlet"
	_ "modserv-shim/internal/shimlet/shimlets"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	exitCode := m.Run()
	os.Exit(exitCode)
}

// 测试正常配置加载流程
func TestGetReg(t *testing.T) {
	a := shimlet.Registry.NewUninitialized("k8s")
	a.InitWithConfig("k8s")
}
