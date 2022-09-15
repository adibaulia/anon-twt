package config

import (
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	timeFormat = "2006-01-02 15:04:05"
)

func initLog() {
	logLevel := conf.LogLevel
	logFilePath := conf.LogPath
	logFileMaxSize := 3
	logFileMaxAge := 3
	logFileMaxBackups := 50
	logFileEnabled := true

	log.SetReportCaller(true)
	log.SetFormatter(&log.JSONFormatter{
		TimestampFormat:  timeFormat,
		DisableTimestamp: false,
	})
	log.SetLevel(getLogLevelFromString(logLevel))

	var fileLogger *lumberjack.Logger
	if logFileEnabled {
		cleanlogFilePath := filepath.Clean(logFilePath)
		dirLogFile := filepath.Dir(cleanlogFilePath)
		if _, err := os.Stat(dirLogFile); errors.Is(err, os.ErrNotExist) {
			err := os.Mkdir(dirLogFile, os.ModePerm)
			if err != nil {
				log.Error(err)
			}
		}
		logFile, err := os.OpenFile(cleanlogFilePath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0600)
		if err != nil {
			log.Errorf("Error opening file, %v. Abort writing log file \n", err)
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
}

// Logger is the logrus logger handler
func Logger(notLogged ...string) gin.HandlerFunc {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	var skip map[string]struct{}

	if length := len(notLogged); length > 0 {
		skip = make(map[string]struct{}, length)

		for _, p := range notLogged {
			skip[p] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		// other handler can change c.Path so:
		path := c.Request.URL.Path
		start := time.Now()
		c.Next()
		stop := time.Since(start)
		latency := int(math.Ceil(float64(stop.Nanoseconds()) / 1000000.0))
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		clientUserAgent := c.Request.UserAgent()
		referer := c.Request.Referer()
		dataLength := c.Writer.Size()
		if dataLength < 0 {
			dataLength = 0
		}

		if _, ok := skip[path]; ok {
			return
		}

		entry := log.WithFields(log.Fields{
			"hostname":   hostname,
			"statusCode": statusCode,
			"latency":    latency, // time to process
			"clientIP":   clientIP,
			"method":     c.Request.Method,
			"path":       path,
			"referer":    referer,
			"dataLength": dataLength,
			"userAgent":  clientUserAgent,
		})

		if len(c.Errors) > 0 {
			entry.Error(c.Errors.ByType(gin.ErrorTypePrivate).String())
		} else {
			msg := fmt.Sprintf("%s - %s [%s] \"%s %s\" %d %d \"%s\" \"%s\" (%dms)", clientIP, hostname, time.Now().Format(timeFormat), c.Request.Method, path, statusCode, dataLength, referer, clientUserAgent, latency)
			if statusCode >= http.StatusInternalServerError {
				entry.Error(msg)
			} else if statusCode >= http.StatusBadRequest {
				entry.Warn(msg)
			} else {
				entry.Info(msg)
			}
		}
	}
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
