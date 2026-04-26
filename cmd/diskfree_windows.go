//go:build windows
// +build windows

package cmd

import (
	"path/filepath"

	"golang.org/x/sys/windows"
)

func DiskFree(path string) uint64 {
	volName := filepath.VolumeName(path)
	//DebugInfo("DiskFree: Drive", volName)

	var freeSpace uint64
	var totalNumberOfBytes uint64
	var totalNumberOfFreeBytes uint64

	err := windows.GetDiskFreeSpaceEx(windows.StringToUTF16Ptr(volName),
		&freeSpace, &totalNumberOfBytes, &totalNumberOfFreeBytes)

	if err != nil {
		DebugInfo("DiskFree: ERROR", err.Error())
		return 0
	}
	return freeSpace
}
