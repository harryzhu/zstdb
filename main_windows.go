//go:build windows
// +build windows

package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
	"zstdb/cmd"

	"golang.org/x/sys/windows"
)

func main() {
	wg := sync.WaitGroup{}
	wg.Add(4)
	go func() {
		cmd.Execute()
	}()

	go func() {
		cmd.BadgerRunValueLogGC()
	}()

	go func() {
		cmd.StartCron()
	}()

	go func() {
		onExit()
	}()

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			WatchDiskFreeSpace()
		}
	}()

	wg.Wait()
}

func onExit() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cmd.StopGrpcServer()
		time.Sleep(time.Second * 2)
		os.Exit(0)
	}()
}

func WatchDiskFreeSpace() error {
	var cmdMinFreeDiskSpaceMB uint64 = cmd.MinFreeDiskSpaceMB
	if cmdMinFreeDiskSpaceMB < 4096 {
		cmdMinFreeDiskSpaceMB = 4096
	}

	var minFreeSpace uint64 = cmdMinFreeDiskSpaceMB << 20

	freeSpace := uint64(0)
	absDataDir, err := filepath.Abs(cmd.DataDir)
	if err != nil {
		DebugInfo("WatchDiskFreeSpace: ERROR", err)
		return err
	}
	absDataDir = filepath.ToSlash(absDataDir)
	if absDataDir != "" {
		freeSpace = DiskFree(absDataDir)
		DebugInfo("Current freespace(MB)", (freeSpace >> 20), ", (≈", (freeSpace >> 30), "GB)")
		DebugInfo("Current threshold(MB)", (minFreeSpace >> 20))
	}

	if freeSpace > 0 && freeSpace < minFreeSpace {
		cmd.IsDisableSet = true
		DebugInfo("Current IsDisableSet", cmd.IsDisableSet)
	}

	if cmd.IsDisableSet == true {
		DebugInfo("zstdb [SET] action had been DISABLED, new data cannot be saved.")
		DebugInfo("Because of the Free Disk Space < ", minFreeSpace)
	}
	return nil
}

func DebugInfo(prefix string, args ...any) {
	if cmd.IsDebug {
		var info []string
		for _, arg := range args {
			info = append(info, fmt.Sprintf("%v", arg))
		}
		log.Printf("INFO: %v: %v\n", prefix, strings.Join(info, ""))
	}
}

func DiskFree(path string) uint64 {
	volName := filepath.VolumeName(path)
	DebugInfo("DiskFree: Drive", volName)

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
