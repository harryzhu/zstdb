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
		DebugInfo("StartCron: --auto-backup-every is using", AutoBackupEvery, ", --auto-backup-dir=", AutoBackupDir)
	}

	ScheduleTask.AddFunc(AutoBackupEvery, func() {
		if AutoBackupDir != "" && AutoBackupEvery != "" {
			DebugInfo("StartCron", AutoBackupEvery, ", Dir: ", AutoBackupDir)
			AutoBackup()
		}
	})
	ScheduleTask.Start()
}
