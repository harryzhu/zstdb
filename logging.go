package sqlconf

import (
	//"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	//"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	ZapLogger           *zap.Logger
	ErrorLogFile        string
	InfoLogFile         string
	CurrentErrorLogFile string
}

func (l *Logger) initLogger(app_logs_dir string, app_name string) *Logger {
	dt := time.Now().Format("20060102")

	MakeDirs(app_logs_dir)

	infoPath := filepath.Join(app_logs_dir, strings.ToLower(strings.Join([]string{app_name, dt, "info.log"}, "_")))
	errPath := filepath.Join(app_logs_dir, strings.ToLower(strings.Join([]string{app_name, dt, "error.log"}, "_")))
	curPath := filepath.Join(app_logs_dir, strings.ToLower(strings.Join([]string{app_name, "current_error.log"}, "_")))

	logger, err := getLogger(infoPath, errPath, curPath)
	if err != nil {
		log.Fatal(err)
	}

	l.ZapLogger = logger
	l.ErrorLogFile = errPath
	l.InfoLogFile = infoPath
	l.CurrentErrorLogFile = curPath

	defer logger.Sync()

	return l
}

func getLogger(infoPath, errorPath, curPath string) (*zap.Logger, error) {
	highPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev >= zap.ErrorLevel
	})

	lowPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev >= zap.DebugLevel
	})

	prodEncoder := zap.NewProductionEncoderConfig()
	prodEncoder.EncodeTime = zapcore.ISO8601TimeEncoder

	curEncoder := zap.NewProductionEncoderConfig()
	curEncoder.EncodeTime = zapcore.ISO8601TimeEncoder

	devEncoder := zap.NewDevelopmentEncoderConfig()
	devEncoder.EncodeTime = zapcore.ISO8601TimeEncoder

	if strings.ToLower(runtime.GOOS) != "windows" {
		devEncoder.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	consoleEncoder := zapcore.NewConsoleEncoder(devEncoder)

	consoleDebugging := zapcore.Lock(os.Stdout)

	lowWriteSyncer, lowClose, err := zap.Open(infoPath)
	if err != nil {
		log.Fatal(err)
		lowClose()
		return nil, err
	}

	highWriteSyncer, highClose, err := zap.Open(errorPath)
	if err != nil {
		log.Fatal(err)
		highClose()
		return nil, err
	}

	if _, e := os.Stat(curPath); e == nil {
		e = os.Remove(curPath)
		if e != nil {
			log.Println(e)
		}
	}

	curWriteSyncer, curClose, err := zap.Open(curPath)
	if err != nil {
		log.Fatal(err)
		curClose()
		return nil, err
	}

	highCore := zapcore.NewCore(zapcore.NewJSONEncoder(prodEncoder), highWriteSyncer, highPriority)
	lowCore := zapcore.NewCore(zapcore.NewJSONEncoder(prodEncoder), lowWriteSyncer, lowPriority)
	curCore := zapcore.NewCore(zapcore.NewJSONEncoder(curEncoder), curWriteSyncer, highPriority)

	consoleCore := zapcore.NewCore(consoleEncoder, consoleDebugging, lowPriority)

	return zap.New(zapcore.NewTee(highCore, lowCore, curCore, consoleCore), zap.AddCaller()), nil
}

func (l *Logger) MailReportCurrentError() error {
	if l.CurrentErrorLogFile == "" {
		return nil
	}

	if fi, err := os.Stat(l.CurrentErrorLogFile); err == nil {
		if fi.Size() > 16 {
			cnt, err := ioutil.ReadFile(l.CurrentErrorLogFile)
			if err == nil {

				mSubject := "[ERROR].[RUNNING]:" + filepath.Base(l.CurrentErrorLogFile)
				mBody := strings.Join([]string{l.CurrentErrorLogFile, "<br/><br/>", "<pre>", string(cnt), "</pre>"}, "")

				Config.Mail.WithMessage(mSubject, mBody).SendMailStartTLS()
			}

		}
	}

	return nil
}
