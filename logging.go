package sqlconf

import (
	//"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	//"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger *zap.Logger
)

func initLogger(app_logs_dir string, app_name string) {
	dt := time.Now().Format("20060102")

	_, err := os.Stat(app_logs_dir)
	if err != nil {
		err := os.MkdirAll(app_logs_dir, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		} else {
			os.Chmod(app_logs_dir, os.ModePerm)
		}
	}

	infoPath := filepath.Join(app_logs_dir, strings.ToLower(strings.Join([]string{app_name, dt, "info.log"}, "_")))
	errPath := filepath.Join(app_logs_dir, strings.ToLower(strings.Join([]string{app_name, dt, "error.log"}, "_")))

	logger, err = getLogger(infoPath, errPath)
	if err != nil {
		log.Fatal(err)
	}

	defer logger.Sync()
}

func getLogger(infoPath, errorPath string) (*zap.Logger, error) {
	highPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev >= zap.ErrorLevel
	})

	lowPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev >= zap.DebugLevel
	})

	prodEncoder := zap.NewProductionEncoderConfig()
	prodEncoder.EncodeTime = zapcore.ISO8601TimeEncoder
	prodEncoder.EncodeLevel = zapcore.CapitalColorLevelEncoder

	devEncoder := zap.NewDevelopmentEncoderConfig()
	devEncoder.EncodeTime = zapcore.ISO8601TimeEncoder
	devEncoder.EncodeLevel = zapcore.CapitalColorLevelEncoder

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

	highCore := zapcore.NewCore(zapcore.NewJSONEncoder(prodEncoder), highWriteSyncer, highPriority)
	lowCore := zapcore.NewCore(zapcore.NewJSONEncoder(prodEncoder), lowWriteSyncer, lowPriority)
	consoleCore := zapcore.NewCore(consoleEncoder, consoleDebugging, lowPriority)

	return zap.New(zapcore.NewTee(highCore, lowCore, consoleCore), zap.AddCaller()), nil
}
