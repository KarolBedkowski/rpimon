package logging

import (
	"io"
	"log"
	"os"
)

var Logger = log.New(os.Stderr, "", log.LstdFlags)
var debugLevel = false

const (
	DEBUG = "DEBUG"
	INFO  = "INFO"
	WARN  = "WARN"
	ERROR = "ERROR"
	FATAL = "FATAL" // die
)

func Init(filename string, debug bool) {
	debugLevel = debug
	outFile, _ := os.Create(filename)

	flags := log.LstdFlags
	if debug {
		flags |= log.Lshortfile
	}
	Logger = log.New(io.MultiWriter(os.Stderr, outFile), "", flags)
}

func Print(v ...interface{}) {
	Logger.Print(v...)
}

func Printf(v ...interface{}) {
	Logger.Print(v...)
}

func Debug(v ...interface{}) {
	if debugLevel {
		Logger.Printf(DEBUG, v...)
	}
}

func Info(v ...interface{}) {
	Logger.Printf(INFO, v...)
}

func Warn(v ...interface{}) {
	Logger.Printf(WARN, v...)
}

func Error(v ...interface{}) {
	Logger.Printf(ERROR, v...)
}
