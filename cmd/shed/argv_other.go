//go:build !windows

package main

func repairPlatformArgs(args []string) []string {
	return args
}

func normalizeFolderArgs(args []string) []string {
	return args
}
