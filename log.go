package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"time"
)

const format string = "[dateTtime] [type] error"

var logDir string = "./"
var logFile *os.File

func init() {
	switch runtime.GOOS {
	case "windows":
		dir := fmt.Sprintf("C:\\ProgramData\\%s\\", appName)
		err := os.Mkdir(dir, 0o755)
		if err != nil {
			logDir = "./"
		}
		logDir = dir
	case "linux", "darwin":
		logDir = "/var/log/"
	}

	var err error
	for {
		logFile, err = os.OpenFile(filepath.Join(logDir, "assemblixserver.log"), os.O_APPEND|os.O_CREATE, 0o644)
		if err != nil {
			logDir = "./"
			continue
		}
		break
	}
}

func logError(err error) {
	if logFile == nil {
		return
	}
	fmt.Fprintln(logFile, parseFormat("ERROR", err))
}
func logWarning(err error) {
	if logFile == nil {
		return
	}
	fmt.Fprintln(logFile, parseFormat("WARNING", err))
}
func logInfo(err error) {
	if logFile == nil {
		return
	}
	fmt.Fprintln(logFile, parseFormat("INFO", err))
}

func parseFormat(logType string, err error) string {
	result := regexp.MustCompile(`(\\)?\bdate\b`).ReplaceAllStringFunc(format, func(match string) string {
		if match[0] == '\\' {
			return match
		}
		return time.Now().Format("2006-01-02")
	})
	result = regexp.MustCompile(`(\\)?\btime\b`).ReplaceAllStringFunc(result, func(match string) string {
		if match[0] == '\\' {
			return match
		}
		return time.Now().Format("15:04:05")
	})
	result = regexp.MustCompile(`(\\)?\btime\b`).ReplaceAllStringFunc(result, func(match string) string {
		if match[0] == '\\' {
			return match
		}
		return logType
	})
	result = regexp.MustCompile(`(\\)?\btime\b`).ReplaceAllStringFunc(result, func(match string) string {
		if match[0] == '\\' {
			return match
		}
		return err.Error()
	})

	return result
}
