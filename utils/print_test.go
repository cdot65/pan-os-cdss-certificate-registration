package utils

import (
	"bytes"
	"github.com/cdot65/pan-os-cdss-certificate-registration/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"testing"
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
		PrintDeviceList(deviceList, logger.New(0, false), false)
	})

	assert.Contains(t, output, "Device List:")
	assert.Contains(t, output, "Device 1:")
	assert.Contains(t, output, "Hostname: device1")
	assert.Contains(t, output, "IP Address: 192.168.1.1")
	assert.Contains(t, output, "Parsed Version: 10.1.0-h1")
}

func TestPrintResults(t *testing.T) {
	results := []string{ // Change this from chan string to []string
		"Device1: Successfully registered WildFire",
		"Device2: Failed to register WildFire",
		"Device3: Successfully registered WildFire",
	}

	output := captureOutput(t, func() {
		PrintResults(results, 3, logger.New(0, false))
	})

	assert.Contains(t, output, "WildFire Registration Results:")
	assert.Contains(t, output, "Device1: Successfully registered WildFire")
	assert.Contains(t, output, "Device2: Failed to register WildFire")
	assert.Contains(t, output, "Device3: Successfully registered WildFire")
}
