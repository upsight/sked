package sked

import (
	"fmt"
	"log"
	"os"
)

// Logger is for logging to a writer. This is not the raft replication log.
type Logger interface {
	Printf(format string, v ...interface{})
	Println(v ...interface{})
	Errorf(format string, v ...interface{})
	Errorln(v ...interface{})
}

// DiscardLogger is a noop logger.
type DiscardLogger struct {
}

// Println noop
func (d *DiscardLogger) Println(v ...interface{}) {}

// Printf noop
func (d *DiscardLogger) Printf(format string, v ...interface{}) {}

// Errorln noop
func (d *DiscardLogger) Errorln(v ...interface{}) {}

// Errorf noop
func (d *DiscardLogger) Errorf(format string, v ...interface{}) {}

// LogLogger uses the std lib logger.
type LogLogger struct {
	logger *log.Logger
}

// NewLogLogger will create a new default logger.
func NewLogLogger() *LogLogger {
	logger := log.New(os.Stderr, "sked: ", log.Lshortfile)
	return &LogLogger{logger}
}

// Println std lib
func (d *LogLogger) Println(v ...interface{}) {
	d.logger.Output(3, "[INFO] "+fmt.Sprintln(v...))
}

// Printf std lib
func (d *LogLogger) Printf(format string, v ...interface{}) {
	d.logger.Output(3, fmt.Sprintf("[INFO] "+format, v...))
}

// Errorln std lib
func (d *LogLogger) Errorln(v ...interface{}) {
	d.logger.Output(3, "[ERRO] "+fmt.Sprintln(v...))
}

// Errorf std lib
func (d *LogLogger) Errorf(format string, v ...interface{}) {
	d.logger.Output(3, fmt.Sprintf("[ERRO] "+format, v...))
}
