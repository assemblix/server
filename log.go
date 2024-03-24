package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"time"
)

var logDir string = "./"
var logFile *os.File

func init() {
	switch runtime.GOOS {
	case "windows":
		dir := "C:\\ProgramData\\assemblix"
		err := os.Mkdir(dir, 0o664)
		if err != nil {
			logWarning(err)
			logDir = "./"
			break
		}
		logDir = dir
	case "linux", "darwin":
		dir := "/var/log/assemblix"
		err := os.Mkdir(dir, 0o664)
		if err != nil {
			logWarning(err)
			logDir = "./"
			break
		}
		logDir = dir
	}

	var err error
	for {
		logFile, err = os.OpenFile(filepath.Join(logDir, "server.log"), os.O_APPEND|os.O_CREATE, 0o644)
		if err != nil {
			logWarning(err)
			logDir = "./"
			continue
		}
		break
	}
}

func logError(err error) {
	var parsed = parseFormat("ERROR", err)
	if logFile == nil {
		fmt.Println(parsed)
		return
	}
	fmt.Fprintln(logFile, parsed)
}
func logWarning(err error) {
	var parsed = parseFormat("WARNING", err)
	if logFile == nil {
		fmt.Println(parsed)
		return
	}
	fmt.Fprintln(logFile, parsed)
}
func logInfo(err error) {
	var parsed = parseFormat("INFO", err)
	if logFile == nil {
		fmt.Println(parsed)
		return
	}
	fmt.Fprintln(logFile, parsed)
}

func parseFormat(logType string, err error) string {
	result := regexp.MustCompile(`(\\)?date`).ReplaceAllStringFunc(logFormat, func(match string) string {
		if match[0] == '\\' {
			return match
		}
		return time.Now().Format("2006-01-02")
	})
	result = regexp.MustCompile(`(\\)?time`).ReplaceAllStringFunc(result, func(match string) string {
		if match[0] == '\\' {
			return match
		}
		return time.Now().Format("15:04:05")
	})
	result = regexp.MustCompile(`(\\)?type`).ReplaceAllStringFunc(result, func(match string) string {
		if match[0] == '\\' {
			return match
		}
		return logType
	})
	result = regexp.MustCompile(`(\\)?error`).ReplaceAllStringFunc(result, func(match string) string {
		if match[0] == '\\' {
			return match
		}
		return err.Error()
	})

	return result
}
