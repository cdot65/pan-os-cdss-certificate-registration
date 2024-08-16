package logger

import (
	"fmt"
	"log"
	"os"
)

// Logger is a custom logger with debug levels.
type Logger struct {
	debugLevel int
	*log.Logger
	exitFunc func(int) // New field for custom exit function
}

// New creates and returns a new Logger instance with specified debug level and verbosity.
func New(debugLevel int, verbose bool) *Logger {
	if verbose {
		debugLevel = 1
	}
	return &Logger{
		debugLevel: debugLevel,
		Logger:     log.New(os.Stdout, "", log.Ldate|log.Ltime),
		exitFunc:   os.Exit, // Default to os.Exit
	}
}

// Debug logs a debug message if the debug level is set to 1 or higher.
func (l *Logger) Debug(v ...interface{}) {
	if l.debugLevel >= 1 {
		l.Printf("[DEBUG] %v", fmt.Sprintln(v...))
	}
}

// Info logs an informational message if the debug level is sufficient.
func (l *Logger) Info(v ...interface{}) {
	if l.debugLevel >= 0 {
		l.Printf("[INFO] %v", fmt.Sprintln(v...))
	}
}

// Fatalf logs a fatal error message and terminates the program.
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.Printf("[FATAL] "+format, v...)
	l.exitFunc(1)
}
