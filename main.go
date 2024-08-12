package main

import (
	"fmt"
	"github.com/cdot65/pan-os-cdss-certificate-registration/config"
	"github.com/cdot65/pan-os-cdss-certificate-registration/devices"
	"github.com/cdot65/pan-os-cdss-certificate-registration/logger"
	"github.com/cdot65/pan-os-cdss-certificate-registration/pdfgenerate"
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
	cfg := config.ParseFlags()

	// Initialize logger
	l := logger.New(cfg.DebugLevel, cfg.Verbose)

	// Load configuration
	conf, err := config.Load(cfg.ConfigFile, cfg.SecretsFile)
	if err != nil {
		l.Fatalf("Failed to load configuration: %v", err)
	}

	// Create DeviceManager
	dm := devices.NewDeviceManager(conf, l)

	// Get device list
	deviceList, err := dm.GetDeviceList(cfg.NoPanorama, cfg.HostnameFilter)
	if err != nil {
		l.Fatalf("Failed to get device list: %v", err)
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
	utils.PrintDeviceList(unaffectedDevices, l, cfg.Verbose)

	// Print message before starting firewall connections
	utils.PrintStartingFirewallConnections(l)

	// Register WildFire for unaffectedDevices devices
	results := make(chan string, len(unaffectedDevices))
	var wg sync.WaitGroup

	for i, device := range unaffectedDevices {
		wg.Add(1)
		go func(dev map[string]string, index int) {
			defer wg.Done()
			err := wildfire.RegisterWildFire(dev, conf.Auth.Auth.Firewall.Username, conf.Auth.Auth.Firewall.Password, l)
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
	var processedResults []string
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

	// Generate PDF report with all information including WildFire registration results
	err = pdfgenerate.GeneratePDFReport(deviceList, affectedDevices, unaffectedDevices, "device_report.pdf")
	if err != nil {
		log.Fatal("Error generating PDF report:", err)
	}

	// Print results
	utils.PrintResults(processedResults, len(unaffectedDevices), l)

}
