package utils

import (
	"fmt"
	"log"
)

type Logger struct {
	context string
}

func NewLogger(context string) *Logger {
	return &Logger{
		context: context,
	}
}

func (l *Logger) Info(format string, a ...interface{}) {
	// Format the message using fmt.Sprintf
	msg := fmt.Sprintf(format, a...)
	// Log the formatted message
	log.Printf("[INFO] %s: %s", l.context, msg)
}

func (l *Logger) Error(context string, err error) {
	log.Printf("[ERROR] %s|%s: %v", context, l.context, err)
}
