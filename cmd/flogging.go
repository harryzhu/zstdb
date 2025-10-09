package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var LogDir string
var LogMaxSizeMB int64
var errorLogFile string
var errorLogFileHandler *os.File
var enableFileLogging bool

var logChan chan string = make(chan string, 1000)

func StartFileLogging() error {
	time.Sleep(1 * time.Second)
	if LogDir == "" {
		return nil
	}

	refreshErrorLogFileHandler()

	go func() {
		var msg string

		fp, err := refreshErrorLogFileHandler()
		if err != nil {
			PrintError("StartFileLogging", err)
			enableFileLogging = false
			return
		}

		for {
			msg = <-logChan
			if msg == "refresh-error-log-file-handler" {
				fp.Close()
				fp, err = refreshErrorLogFileHandler()
				if err != nil {
					PrintError("refreshErrorLogFileHandler", err)
					enableFileLogging = false
					break
				}
				continue
			}
			if msg != "" {
				fp.WriteString(fmt.Sprintf("%v\n", msg))
			}

		}
	}()

	return nil
}

func StopFileLogging() {
	if LogDir != "" {
		return
	}
	enableFileLogging = false
	close(logChan)
	errorLogFileHandler.Close()
}

func refreshErrorLogFileHandler() (*os.File, error) {
	if LogDir == "" {
		return nil, NewError("file logging is not active")
	}

	enableFileLogging = false
	if errorLogFile != "" {
		errorLogFileHandler.Close()
	}
	errorLogFile = ToUnixSlash(filepath.Join(LogDir, strings.Join([]string{"zstdb", Host, Port, time.Now().Format("20060102"), "error.log"}, "_")))
	if LogMaxSizeMB < 1 {
		LogMaxSizeMB = 1
	}
	lfinfo, err := os.Stat(errorLogFile)
	if err == nil {
		maxSize := LogMaxSizeMB << 20
		if lfinfo.Size() > maxSize {
			err = os.Truncate(errorLogFile, 0)
			DebugWarn("refreshErrorLogFileHandler", "truncate log file")
			if err != nil {
				PrintError("refreshErrorLogFileHandler", err)
				return nil, err
			}
		}
	}

	errorLogFileHandler, err = os.OpenFile(errorLogFile, os.O_CREATE|os.O_RDWR|os.O_APPEND, os.ModeAppend|os.ModePerm)
	if err != nil {
		PrintError("cannot open log file", err)
		return nil, err
	}
	enableFileLogging = true

	return errorLogFileHandler, nil
}

func WatchErrorLogFile() error {
	if LogDir == "" {
		return nil
	}

	lfinfo, err := os.Stat(errorLogFile)
	if err != nil {
		logChan <- "refresh-error-log-file-handler"
		PrintError("WatchErrorLogFile", err)
		DebugWarn("WatchErrorLogFile", errorLogFile)
		return err
	}
	maxSize := LogMaxSizeMB << 20
	if lfinfo.Size() > maxSize {
		logChan <- "refresh-error-log-file-handler"
	}
	if time.Now().Format("1504") == "0000" {
		logChan <- "refresh-error-log-file-handler"
	}

	return nil
}

func flog(args ...any) error {
	if LogDir == "" || !enableFileLogging {
		return nil
	}

	var info []string
	info = append(info, time.Now().Format("2006-01-02 15:04:05"))
	for _, arg := range args {
		info = append(info, fmt.Sprintf("%v", arg))
	}
	logChan <- strings.Join(info, " ")
	return nil
}
