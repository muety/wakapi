package utils

import (
	"math"
	"runtime"
)

func AllCPUs() int {
	return runtime.NumCPU()
}

func HalfCPUs() int {
	return int(math.Ceil(float64(runtime.NumCPU()) / 2.0))
}
