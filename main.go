// Package main.go
package main

import (
	"fmt"
	"github.com/cdot65/pan-os-cdss-certificate-registration/config"
	"github.com/cdot65/pan-os-cdss-certificate-registration/devices"
	"github.com/cdot65/pan-os-cdss-certificate-registration/logger"
	"github.com/cdot65/pan-os-cdss-certificate-registration/utils/consoleprint"
	"github.com/cdot65/pan-os-cdss-certificate-registration/utils/filters"
	"github.com/cdot65/pan-os-cdss-certificate-registration/utils/pdf"
	"github.com/cdot65/pan-os-cdss-certificate-registration/wildfire"
	"log"
	"strings"
	"sync"
)

// Main function to register WildFire on multiple devices concurrently.
// This function parses command-line flags, loads configuration, retrieves a list of devices,
// and concurrently registers WildFire on each device. It uses goroutines for parallel processing
// and reports the results for each device.
func main() {
	// Parse command-line flags
	flags, _ := config.ParseFlags()

	// Initialize logger
	l := logger.New(flags.DebugLevel, flags.Verbose)

	// Load configuration
	conf, err := config.Load(flags.ConfigFile, flags.SecretsFile, flags)
	if err != nil {
		l.Fatalf("Failed to load configuration: %v", err)
	}

	// Create DeviceManager
	dm := devices.NewDeviceManager(conf, l)

	// Get device list
	deviceList, err := dm.GetDeviceList(flags.NoPanorama)
	if err != nil {
		l.Fatalf("Failed to get device list: %v", err)
	}

	// Check if we got any devices
	if len(deviceList) == 0 {
		l.Fatalf("No devices were successfully processed")
	}

	// Filter devices by hardware family
	eligibleHardware, ineligibleHardware := filters.FilterDevicesByFamily(deviceList)

	// Parse versions and update eligibleHardware
	for i, device := range eligibleHardware {
		swVersion := device["sw-version"]
		parsedVersion, err := filters.ParseVersion(swVersion)
		if err != nil {
			l.Fatalf("Failed to parse version for device %s: %v", device["hostname"], err)
		}

		// Add parsed version components to the device map
		eligibleHardware[i]["parsed_version_major"] = fmt.Sprintf("%d", parsedVersion.Major)
		eligibleHardware[i]["parsed_version_feature"] = fmt.Sprintf("%d", parsedVersion.Feature)
		eligibleHardware[i]["parsed_version_maintenance"] = fmt.Sprintf("%d", parsedVersion.Maintenance)
		eligibleHardware[i]["parsed_version_hotfix"] = fmt.Sprintf("%d", parsedVersion.Hotfix)
	}

	// Split eligible hardware devices into supported and unsupported versions
	supportedVersions, unsupportedVersions, err := filters.SplitDevicesByVersion(eligibleHardware)
	if err != nil {
		l.Fatalf("Failed to split devices by version: %v", err)
	}

	// The registrationCandidates are the devices with supported versions
	registrationCandidates := supportedVersions

	// Print registration candidates list
	consoleprint.PrintDeviceList(registrationCandidates, l, flags.Verbose)

	// Print message before starting firewall connections
	consoleprint.PrintStartingFirewallConnections(l)

	var processedResults []string

	if !flags.ReportOnly {
		// Register WildFire for registration candidates
		results := make(chan string, len(registrationCandidates))
		var wg sync.WaitGroup

		for i, device := range registrationCandidates {
			wg.Add(1)
			go func(dev map[string]string, index int) {
				defer wg.Done()
				err := wildfire.RegisterWildFire(dev, conf.Auth.Credentials.Firewall.Username, conf.Auth.Credentials.Firewall.Password, l)
				if err != nil {
					results <- fmt.Sprintf("%s: Failed to register WildFire - %v", dev["hostname"], err)
				} else {
					results <- fmt.Sprintf("%s: Successfully registered WildFire", dev["hostname"])
				}
			}(device, i)
		}

		// Wait for all goroutines to finish
		wg.Wait()
		close(results)

		// Process results and update registrationCandidates
		for result := range results {
			processedResults = append(processedResults, result)
			parts := strings.SplitN(result, ": ", 2)
			if len(parts) == 2 {
				hostname, resultText := parts[0], parts[1]
				for i, device := range registrationCandidates {
					if device["hostname"] == hostname {
						registrationCandidates[i]["result"] = resultText
						break
					}
				}
			}
		}
	} else {
		// Report-only mode: Set a message for registration candidates
		for i := range registrationCandidates {
			registrationCandidates[i]["result"] = "Skipped WildFire registration (Report-only mode)"
		}
	}

	// Generate PDF report
	err = pdf.GeneratePDFReport(deviceList, ineligibleHardware, unsupportedVersions, registrationCandidates, "device_report.pdf")
	if err != nil {
		log.Fatal("Error generating PDF report:", err)
	}

	// Print results
	consoleprint.PrintResults(processedResults, len(registrationCandidates), l)
}
