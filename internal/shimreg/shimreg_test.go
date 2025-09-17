package shimreg_test

import (
	"modserv-shim/internal/shimreg"
	_ "modserv-shim/internal/shimreg/shimlets"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	exitCode := m.Run()
	os.Exit(exitCode)
}

// 测试正常配置加载流程
func TestGetReg(t *testing.T) {
	a := shimreg.NewUninitialized("k8s")
	a.InitWithConfig("k8s")
}
