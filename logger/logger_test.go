package logger

import (
	"bytes"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("Default debug level", func(t *testing.T) {
		logger := New(0, false)
		assert.Equal(t, 0, logger.debugLevel)
	})

	t.Run("Verbose mode", func(t *testing.T) {
		logger := New(0, true)
		assert.Equal(t, 1, logger.debugLevel)
	})

	t.Run("Custom debug level", func(t *testing.T) {
		logger := New(2, false)
		assert.Equal(t, 2, logger.debugLevel)
	})
}

func TestDebug(t *testing.T) {
	var buf bytes.Buffer

	t.Run("Debug level 0", func(t *testing.T) {
		logger := &Logger{debugLevel: 0, Logger: log.New(&buf, "", 0)}
		logger.Debug("This is a debug message")
		assert.Empty(t, buf.String())
	})

	t.Run("Debug level 1", func(t *testing.T) {
		buf.Reset()
		logger := &Logger{debugLevel: 1, Logger: log.New(&buf, "", 0)}
		logger.Debug("This is a debug message")
		assert.Contains(t, buf.String(), "[DEBUG] This is a debug message")
	})
}

func TestInfo(t *testing.T) {
	var buf bytes.Buffer

	t.Run("Debug level 0", func(t *testing.T) {
		logger := &Logger{debugLevel: 0, Logger: log.New(&buf, "", 0)}
		logger.Info("This is an info message")
		assert.Contains(t, buf.String(), "[INFO] This is an info message")
	})

	t.Run("Debug level 1", func(t *testing.T) {
		buf.Reset()
		logger := &Logger{debugLevel: 1, Logger: log.New(&buf, "", 0)}
		logger.Info("This is an info message")
		assert.Contains(t, buf.String(), "[INFO] This is an info message")
	})
}

func TestFatalf(t *testing.T) {
	var buf bytes.Buffer
	var exitCode int

	logger := &Logger{
		debugLevel: 0,
		Logger:     log.New(&buf, "", 0),
		exitFunc: func(code int) {
			exitCode = code
		},
	}

	logger.Fatalf("This is a fatal error: %s", "test")

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, buf.String(), "[FATAL] This is a fatal error: test")
}
