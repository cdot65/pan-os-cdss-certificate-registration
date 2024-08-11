package utils

import (
	"fmt"
	"strings"
	"time"

	"github.com/cdot65/pan-os-cdss-certificate-registration/logger"
)

// PrintDeviceList prints a formatted list of devices to the console.
func PrintDeviceList(deviceList []map[string]string, l *logger.Logger) {
	l.Info("Printing device list")
	fmt.Println("Device List:")
	for i, device := range deviceList {
		fmt.Printf("Device %d:\n", i+1)
		for key, value := range device {
			fmt.Printf("  %s: %s\n", key, value)
		}
		fmt.Printf("  Parsed Version: %s.%s.%s-h%s\n",
			device["parsed_version_major"],
			device["parsed_version_feature"],
			device["parsed_version_maintenance"],
			device["parsed_version_hotfix"])
		fmt.Println()
	}
}

// PrintResults processes and displays WildFire registration results for multiple devices.
func PrintResults(results chan string, totalDevices int, l *logger.Logger) {
	l.Info("Waiting for WildFire registration results")
	fmt.Println("WildFire Registration Results:")
	successCount := 0
	failureCount := 0
	for i := 0; i < totalDevices; i++ {
		select {
		case result, ok := <-results:
			if !ok {
				l.Info("Results channel closed unexpectedly")
				break
			}
			fmt.Println(result)
			if strings.Contains(result, "Successfully registered") {
				successCount++
			} else {
				failureCount++
			}
		case <-time.After(6 * time.Minute):
			l.Info("Timeout waiting for result")
			fmt.Printf("Timeout waiting for result from device %d\n", i+1)
			failureCount++
		}
	}

	l.Info(fmt.Sprintf("Registration complete. Successes: %d, Failures: %d", successCount, failureCount))
}
