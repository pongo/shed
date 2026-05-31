package main

import "runtime"

func unsupportedPlatform() bool {
	return runtime.GOOS != "windows"
}
