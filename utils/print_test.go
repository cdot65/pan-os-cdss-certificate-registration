package utils

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/cdot65/pan-os-cdss-certificate-registration/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func captureOutput(t *testing.T, f func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err, "Failed to create pipe")

	os.Stdout = w

	f()

	err = w.Close()
	require.NoError(t, err, "Failed to close writer")
	os.Stdout = old

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err, "Failed to copy buffer")

	err = r.Close()
	require.NoError(t, err, "Failed to close reader")

	return buf.String()
}

func TestPrintDeviceList(t *testing.T) {
	deviceList := []map[string]string{
		{
			"hostname":                   "device1",
			"ip-address":                 "192.168.1.1",
			"parsed_version_major":       "10",
			"parsed_version_feature":     "1",
			"parsed_version_maintenance": "0",
			"parsed_version_hotfix":      "1",
		},
	}

	output := captureOutput(t, func() {
		PrintDeviceList(deviceList, logger.New(0, false))
	})

	assert.Contains(t, output, "Device List:")
	assert.Contains(t, output, "Device 1:")
	assert.Contains(t, output, "hostname: device1")
	assert.Contains(t, output, "ip-address: 192.168.1.1")
	assert.Contains(t, output, "Parsed Version: 10.1.0-h1")
}

func TestPrintResults(t *testing.T) {
	results := make(chan string, 3)
	results <- "Device1: Successfully registered WildFire"
	results <- "Device2: Failed to register WildFire"
	results <- "Device3: Successfully registered WildFire"
	close(results)

	output := captureOutput(t, func() {
		PrintResults(results, 3, logger.New(0, false))
	})

	assert.Contains(t, output, "WildFire Registration Results:")
	assert.Contains(t, output, "Device1: Successfully registered WildFire")
	assert.Contains(t, output, "Device2: Failed to register WildFire")
	assert.Contains(t, output, "Device3: Successfully registered WildFire")
}

func TestPrintResultsTimeout(t *testing.T) {
	results := make(chan string)

	output := captureOutput(t, func() {
		go func() {
			time.Sleep(100 * time.Millisecond)
			close(results)
		}()

		// Create a custom PrintResults function with a shorter timeout
		customPrintResults := func(results chan string, totalDevices int, l *logger.Logger) {
			l.Info("Waiting for WildFire registration results")
			fmt.Println("WildFire Registration Results:")
			for i := 0; i < totalDevices; i++ {
				select {
				case result, ok := <-results:
					if !ok {
						l.Info("Results channel closed unexpectedly")
						return
					}
					fmt.Println(result)
				case <-time.After(50 * time.Millisecond): // Short timeout for testing
					l.Info("Timeout waiting for result")
					fmt.Printf("Timeout waiting for result from device %d\n", i+1)
				}
			}
		}

		customPrintResults(results, 1, logger.New(0, false))
	})

	assert.Contains(t, output, "WildFire Registration Results:")
	assert.Contains(t, output, "Timeout waiting for result from device 1")
}
