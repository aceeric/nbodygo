package logging

import (
	"github.com/hashicorp/logutils"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	// the environment variable that passes the log level to the app
	logEnvVar = "NBODYLOG"
)

// Valid logging levels
var levels = []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR", "OFF"}

//
// Initializes the Hashicorp logging filter from the environment, and applies the filter to the
// built-in Go logging capability
//
func InitializeLogging() {
	writer := os.Stderr
	minLevel := toLogLevel(os.Getenv(logEnvVar))
	if minLevel != "OFF" {
		writer = initLogWriter()
	}
	filter := &logutils.LevelFilter{
		Levels:   levels,
		MinLevel: minLevel,
		Writer:   writer,
	}
	log.SetOutput(filter)
}

//
// Translates the passed string to a Hashicorp LogLevel. If empty, returns "OFF". If invalid,
// returns "OFF". Therefore running the app without the NBODYLOG env var defined, or with 'NBODYLOG=',
// or with 'NBODYLOG=UNKNOWN' results in no logging of log statements containing log filters. (Non-
// filtered log statements always log to stderr)
//
func toLogLevel(level string) logutils.LogLevel {
	l := logutils.LogLevel(strings.ToUpper(level))
	for _, lvl := range levels {
		if l == lvl {
			return l
		}
	}
	return "OFF"
}

//
// Creates/opens a log file for append. The log file is '/foo/bar/log/log' where
// '/foo/bar' is the directory in which the app is running, and 'log' is a directory
// under '/foo/bar', and 'log' is a file in directory '/foo/bar/log'.
//
// returns: a pointer to the opened/created 'File', or exits via log.Fatal if there is any
// error
//
func initLogWriter() *os.File {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal("Logging: unable to determine logging directory\n")
	}
	dir += "/log"
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		log.Fatalf("Logging: unable to create logging directory: %v\n", dir)
	}
	logFile := dir + "/log"
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Logging: error creating/opening log file: %v\n", logFile)
	}
	return f
}
