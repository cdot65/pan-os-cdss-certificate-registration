package main

import (
	"fmt"
	"github.com/cdot65/pan-os-cdss-certificate-registration/config"
	"github.com/cdot65/pan-os-cdss-certificate-registration/devices"
	"github.com/cdot65/pan-os-cdss-certificate-registration/logger"
	"github.com/cdot65/pan-os-cdss-certificate-registration/utils"
	"github.com/cdot65/pan-os-cdss-certificate-registration/wildfire"
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

	// Get device list
	deviceList, err := devices.GetDeviceList(conf, cfg.NoPanorama, cfg.HostnameFilter, l)
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

	// Filter affected devices
	affectedDevices, err := utils.FilterAffectedDevices(deviceList)
	if err != nil {
		l.Fatalf("Failed to filter affected devices: %v", err)
	}

	// Print affected device list
	utils.PrintDeviceList(affectedDevices, l)

	// Register WildFire
	results := make(chan string, len(affectedDevices))
	var wg sync.WaitGroup

	for i, device := range affectedDevices {
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

	// Close the results channel when all goroutines are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Print results
	utils.PrintResults(results, len(affectedDevices), l)
}
