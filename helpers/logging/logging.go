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
	Logger = log.New(io.MultiWriter(os.Stderr, outFile), "", log.LstdFlags)
}

func Print(v ...interface{}) {
	Logger.Print(v...)
}

func Printf(v ...interface{}) {
	Logger.Print(v...)
}

func Debug(v ...interface{}) {
	if debugLevel {
		Logger.Printf(DEBUG+" "+v[0].(string), v[1:]...)
	}
}

func Info(v ...interface{}) {
	Logger.Printf(INFO+" "+v[0].(string), v[1:]...)
}

func Warn(v ...interface{}) {
	Logger.Printf(WARN+" "+v[0].(string), v[1:]...)
}

func Error(v ...interface{}) {
	Logger.Printf(ERROR+" "+v[0].(string), v[1:]...)
}
