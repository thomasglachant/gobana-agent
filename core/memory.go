package core

import (
	"fmt"
	"runtime"
)

func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Memory usage = %v MiB\n", bToMb(m.Sys))
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024 //nolint:gomnd
}
