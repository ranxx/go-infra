package utils

import (
	"os"
	"runtime/debug"
)

// CurrentGStack 堆栈信息
func CurrentGStack() string {
	data := debug.Stack()
	os.Stderr.Write(data)
	return string(data)
}
