package logging

import (
	"io"
	"log"
	"os"
)

var logger = log.New(os.Stderr, "", log.LstdFlags)
var debugLevel = false

const (
	// DEBUG message prefix
	DEBUG = "DEBUG"
	// INFO message prefix
	INFO = "INFO"
	// WARN message  prefix
	WARN = "WARN"
	// ERROR message prefix
	ERROR = "ERROR"
	// FATAL level prefix
	FATAL = "FATAL" // die
)

// Init logging
func Init(filename string, debug bool) {
	log.Printf("Logging to %s\n", filename)
	debugLevel = debug
	outFile, _ := os.Create(filename)
	logger = log.New(io.MultiWriter(os.Stderr, outFile), "", log.LstdFlags)
}

// Print - wrapper on logger.Print
func Print(v ...interface{}) {
	logger.Print(v...)
}

// Printf - wrapper on logger.Print
func Printf(format string, v ...interface{}) {
	logger.Printf(format, v...)
}

// Debug display message with "DEBUG" prefix when debug=true
func Debug(v ...interface{}) {
	if debugLevel {
		logger.Printf(DEBUG+" "+v[0].(string), v[1:]...)
	}
}

// Info display message with "INFO" prefix
func Info(v ...interface{}) {
	logger.Printf(INFO+" "+v[0].(string), v[1:]...)
}

// Warn display message with "WARN" prefix
func Warn(v ...interface{}) {
	logger.Printf(WARN+" "+v[0].(string), v[1:]...)
}

// Error display message with "ERROR" prefix
func Error(v ...interface{}) {
	logger.Printf(ERROR+" "+v[0].(string), v[1:]...)
}
