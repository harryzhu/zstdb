//go:build !windows
// +build !windows

package cmd

import (
	"syscall"
)

func DiskFree(path string) uint64 {
	fs := syscall.Statfs_t{}
	err := syscall.Statfs(path, &fs)
	if err != nil {
		DebugInfo("DiskFree: ERROR", err)
		return 0
	}
	freeSpace := fs.Bfree * uint64(fs.Bsize)
	return freeSpace
}
