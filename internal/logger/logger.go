package logger

import "log"

type Logger struct {
	context string
}

func NewLogger(context string) *Logger {
	return &Logger{context: context}
}

func (l Logger) Info(message string, args ...interface{}) {
	log.Printf("[INFO] [%s] "+message, append([]interface{}{l.context}, args...)...)
}
func (l Logger) Error(message string, args ...interface{}) {
	log.Printf("[ERROR] [%s] "+message, append([]interface{}{l.context}, args...)...)
}

func (l Logger) Debug(message string, args ...interface{}) {
	log.Printf("[DEBUG] [%s] "+message, append([]interface{}{l.context}, args...)...)
}

func (l Logger) Fatal(message string, args ...interface{}) {
	log.Fatalf("[FATAL] [%s] "+message, append([]interface{}{l.context}, args...)...)
}

func (l Logger) Panic(message string, args ...interface{}) {
	log.Panicf("[PANIC] [%s] "+message, append([]interface{}{l.context}, args...)...)
}

func (l Logger) Warn(message string, args ...interface{}) {
	log.Printf("[WARN] [%s] "+message, append([]interface{}{l.context}, args...)...)
}
