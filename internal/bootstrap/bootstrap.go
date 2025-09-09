package bootstrap

import (
	"sync"
)

var (
	wg sync.WaitGroup
)

func Init(configPath string) error {
	// TODO bootstrap steps impl

	return nil
}

// registerShutdownHook
func registerShutdownHook() {
	// TODO shutdown hook impl

}

func WaitForShutDown() {
}
