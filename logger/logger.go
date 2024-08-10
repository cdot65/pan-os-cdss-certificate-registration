// Package logger/logger.go
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
	exitFunc func(int)
}

// New creates and returns a new Logger instance with specified debug level and verbosity.

// This function initializes a Logger with the given debug level, which can be overridden
// if verbose mode is enabled. It uses the standard log package for output.

// Attributes:
//   debugLevel (int): The initial debug level for the logger.
//   verbose (bool): If true, sets the debug level to 1 regardless of the input debugLevel.

// Return:
//   *Logger: A pointer to the newly created Logger instance.

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

// This method checks the current debug level and, if it's 1 or higher,
// logs the provided message with a "[DEBUG]" prefix using the logger's
// Printf method.

// Attributes:
//   v (...interface{}): Variable number of arguments to be logged.

// Return:
//   None

func (l *Logger) Debug(v ...interface{}) {
	if l.debugLevel >= 1 {
		l.Printf("[DEBUG] %v", fmt.Sprintln(v...))
	}
}

// Info logs an informational message if the debug level is sufficient.

// This method logs a message with the "[INFO]" prefix if the logger's debug level
// is greater than or equal to 0. It uses fmt.Sprintln to format the variadic arguments.

// Attributes:
//   v (...interface{}): Variadic arguments to be logged.

// Return:
//   None

func (l *Logger) Info(v ...interface{}) {
	if l.debugLevel >= 0 {
		l.Printf("[INFO] %v", fmt.Sprintln(v...))
	}
}

// Fatalf logs a fatal error message and terminates the program.

// This method logs a formatted error message with a "[FATAL]" prefix using
// the logger's Printf method, then terminates the program with an exit code of 1.

// Attributes:
//   format (string): The format string for the error message.
//   v (...interface{}): Variable number of arguments to be formatted.

// Error:
//   None explicitly, but the program terminates after logging.

// Return:
//   None (the function does not return due to os.Exit(1)).

func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.Printf("[FATAL] "+format, v...)
	l.exitFunc(1) // Use the exitFunc instead of directly calling os.Exit
}
