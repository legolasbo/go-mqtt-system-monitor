package main

import (
	"log"
)

type Logger struct {
	Config Config
}

type LogLevel int

const TRACE = LogLevel(4)
const DEBUG = LogLevel(3)
const INFO = LogLevel(2)
const WARN = LogLevel(1)
const ERROR = LogLevel(0)

func (l *Logger) Trace(msg string) {
	if l.Config.Verbosity >= TRACE {
		log.Printf("[TRACE]: %s", msg)
	}
}
func (l *Logger) Debug(msg string) {
	if l.Config.Verbosity >= DEBUG {
		log.Printf("[DEBUG]: %s", msg)
	}
}
func (l *Logger) Info(msg string) {
	if l.Config.Verbosity >= INFO {
		log.Printf("[INFO]: %s", msg)
	}
}
func (l *Logger) Warn(msg string) {
	if l.Config.Verbosity >= WARN {
		log.Printf("[WARN]: %s", msg)
	}
}
func (l *Logger) Error(msg string) {
	if l.Config.Verbosity >= ERROR {
		log.Printf("[ERROR]: %s", msg)
	}
}

func createLogger(config Config) Logger {
	return Logger{Config: config}
}
