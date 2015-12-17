package logging

import (
	"io"
	"log"
	"os"
)

var (
	logger     = log.New(os.Stderr, "", log.LstdFlags)
	debugLevel = false
)

const (
	// DEBUG message prefix
	DEBUG = "DEBUG "
	// INFO message prefix
	INFO = "INFO  "
	// WARN message  prefix
	WARN = "WARN  "
	// ERROR message prefix
	ERROR = "ERROR "
	// FATAL level prefix
	FATAL = "FATAL " // die
)

// Init logging
func Init(filename string, debug bool) {
	log.Printf("Logging to %s\n", filename)
	debugLevel = debug
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND, 0660)
	if err != nil {
		f, err = os.Create(filename)
	}
	if err != nil {
		log.Printf("Opening %s for writting error %s\n", filename, err.Error())
	} else {
		logger = log.New(io.MultiWriter(os.Stderr, f), "", log.LstdFlags)
	}
}

// Debug display message with "DEBUG" prefix when debug=true
func Debug(format string, v ...interface{}) {
	if debugLevel {
		logger.Printf(DEBUG+format, v...)
	}
}

// Info display message with "INFO" prefix
func Info(format string, v ...interface{}) {
	logger.Printf(INFO+format, v...)
}

// Warn display message with "WARN" prefix
func Warn(format string, v ...interface{}) {
	logger.Printf(WARN+format, v...)
}

// Error display message with "ERROR" prefix
func Error(format string, v ...interface{}) {
	logger.Printf(ERROR+format, v...)
}
