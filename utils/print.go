package utils

import (
	"fmt"
	"github.com/cdot65/pan-os-cdss-certificate-registration/logger"
	"strings"
)

func PrintDeviceList(deviceList []map[string]string, l *logger.Logger, verbose bool) {
	l.Info("Printing device list")
	fmt.Println("Device List:")
	for i, device := range deviceList {
		fmt.Printf("Device %d:\n", i+1)
		if verbose {
			for key, value := range device {
				fmt.Printf("  %s: %s\n", key, value)
			}
		} else {
			fmt.Printf("  Hostname: %s\n", device["hostname"])
			fmt.Printf("  IP Address: %s\n", device["ip-address"])
			fmt.Printf("  Parsed Version: %s.%s.%s-h%s\n",
				device["parsed_version_major"],
				device["parsed_version_feature"],
				device["parsed_version_maintenance"],
				device["parsed_version_hotfix"])
		}
		fmt.Println()
	}
}

// PrintResults processes and displays WildFire registration results for multiple devices.
func PrintResults(results []string, totalDevices int, l *logger.Logger) {
	l.Info("Processing WildFire registration results")
	fmt.Println("WildFire Registration Results:")
	successCount := 0
	failureCount := 0

	for _, result := range results {
		fmt.Println(result)
		if strings.Contains(result, "Successfully registered") {
			successCount++
		} else {
			failureCount++
		}
	}

	// Check if we have results for all devices
	if len(results) < totalDevices {
		missingResults := totalDevices - len(results)
		l.Info(fmt.Sprintf("Missing results for %d device(s)", missingResults))
		failureCount += missingResults
	}

	l.Info(fmt.Sprintf("Registration complete. Successes: %d, Failures: %d", successCount, failureCount))
}

func PrintStartingFirewallConnections(l *logger.Logger) {
	l.Info("Starting connections to firewalls using scrapli-go")
	fmt.Println("Initiating connections to firewalls for WildFire registration...")
}
