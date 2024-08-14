package main

import (
	"fmt"
	"github.com/cdot65/pan-os-cdss-certificate-registration/config"
	"github.com/cdot65/pan-os-cdss-certificate-registration/devices"
	"github.com/cdot65/pan-os-cdss-certificate-registration/logger"
	"github.com/cdot65/pan-os-cdss-certificate-registration/utils"
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

	// Parse versions and update deviceList
	for i, device := range deviceList {
		swVersion := device["sw-version"]
		parsedVersion, err := utils.ParseVersion(swVersion)
		if err != nil {
			l.Fatalf("Failed to parse version for device %s: %v", device["hostname"], err)
		}

		// Add parsed version components to the device map
		deviceList[i]["parsed_version_major"] = fmt.Sprintf("%d", parsedVersion.Major)
		deviceList[i]["parsed_version_feature"] = fmt.Sprintf("%d", parsedVersion.Feature)
		deviceList[i]["parsed_version_maintenance"] = fmt.Sprintf("%d", parsedVersion.Maintenance)
		deviceList[i]["parsed_version_hotfix"] = fmt.Sprintf("%d", parsedVersion.Hotfix)
	}

	// Split devices into affected and unaffected
	affectedDevices, unaffectedDevices, err := utils.SplitDevices(deviceList)
	if err != nil {
		l.Fatalf("Failed to split devices: %v", err)
	}

	// Print unaffectedDevices device list
	utils.PrintDeviceList(unaffectedDevices, l, flags.Verbose)

	// Print message before starting firewall connections
	utils.PrintStartingFirewallConnections(l)

	// Create an empty placeholder that will contain the result our Scrapli tasks
	var processedResults []string

	// If we are not running in reportonly mode, then construct channels and a WaitGroup to safely concurrently connect
	if !flags.ReportOnly {
		// Register WildFire for unaffectedDevices devices
		results := make(chan string, len(unaffectedDevices))
		var wg sync.WaitGroup

		for i, device := range unaffectedDevices {
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

		// Close the results channel
		close(results)

		// Process results and update unaffectedDevices
		for result := range results {
			processedResults = append(processedResults, result)
			parts := strings.SplitN(result, ": ", 2)
			if len(parts) == 2 {
				hostname, resultText := parts[0], parts[1]
				for i, device := range unaffectedDevices {
					if device["hostname"] == hostname {
						unaffectedDevices[i]["result"] = resultText
						break
					}
				}
			}
		}
	} else {
		// Report-only mode: Set a message for unaffected devices
		for i := range unaffectedDevices {
			unaffectedDevices[i]["result"] = "Skipped WildFire registration (Report-only mode)"
		}
	}

	// Generate PDF report with all information including WildFire registration results
	err = utils.GeneratePDFReport(deviceList, affectedDevices, unaffectedDevices, "device_report.pdf")
	if err != nil {
		log.Fatal("Error generating PDF report:", err)
	}

	// Print results
	utils.PrintResults(processedResults, len(unaffectedDevices), l)
}
