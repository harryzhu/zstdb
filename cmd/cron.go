package cmd

import (
	"time"

	"github.com/robfig/cron/v3"
)

var (
	ScheduleTask *cron.Cron
)

func StartCron() {
	ScheduleTask = cron.New()
	time.Sleep(3 * time.Second)
	if AutoBackupEvery != "" {
		_, err := cron.ParseStandard(AutoBackupEvery)
		if err != nil {
			PrintError("StartCron,--auto-backup-every= invalid, will use default", err)
			AutoBackupEvery = "@every 1h"
		}

	}

	if AutoBackupDir != "" && AutoBackupEvery != "" {
		DebugInfo("StartCron: --auto-backup-every is using", AutoBackupEvery, ", --auto-backup-dir=", AutoBackupDir)
	} else {
		DebugInfo("StartCron: AutoBackup", "disabled")
	}

	ScheduleTask.AddFunc(AutoBackupEvery, func() {
		if AutoBackupDir != "" && AutoBackupEvery != "" {
			DebugInfo("StartCron", AutoBackupEvery, ", Dir: ", AutoBackupDir)
			AutoBackup()
		}
	})

	if enableFileLogging {
		DebugInfo("StartCron: FileLogging", "enabled => ", errorLogFile)
	} else {
		DebugInfo("StartCron: FileLogging", "disabled")
	}

	ScheduleTask.AddFunc("@every 1m", func() {
		WatchErrorLogFile()
	})

	ScheduleTask.Start()
}
