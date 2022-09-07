package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

func initLog() {
	logLevel := conf.logLevel
	logFilePath := conf.logPath
	logFileMaxSize := 3
	logFileMaxAge := 3
	logFileMaxBackups := 50
	logFileEnabled := true

	var fileLogger *lumberjack.Logger

	if logFileEnabled {
		cleanlogFilePath := filepath.Clean(logFilePath)
		dirLogFile := filepath.Dir(cleanlogFilePath)
		if _, err := os.Stat(dirLogFile); errors.Is(err, os.ErrNotExist) {
			err := os.Mkdir(dirLogFile, os.ModePerm)
			if err != nil {
				log.Println(err)
			}
		}
		logFile, err := os.OpenFile(cleanlogFilePath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0600)
		if err != nil {
			fmt.Printf("Error opening file, %v. Abort writing log file \n", err)
			return
		}

		fileLogger = &lumberjack.Logger{
			Filename:   logFile.Name(),
			MaxSize:    logFileMaxSize,
			MaxAge:     logFileMaxAge,
			MaxBackups: logFileMaxBackups,
			LocalTime:  true,
			Compress:   true,
		}
	}

	mWriter := io.MultiWriter(os.Stdout, fileLogger)

	log.SetOutput(mWriter)
	log.SetReportCaller(true)
	log.SetFormatter(&log.JSONFormatter{
		TimestampFormat:  "2006-01-02 15:04:05",
		DisableTimestamp: false,
	})
	log.SetLevel(getLogLevelFromString(logLevel))
}

func getLogLevelFromString(level string) log.Level {
	switch strings.ToLower(level) {
	case "warn":
		return log.WarnLevel
	case "debug":
		return log.DebugLevel
	case "error":
		return log.ErrorLevel
	default:
		return log.InfoLevel
	}
}
