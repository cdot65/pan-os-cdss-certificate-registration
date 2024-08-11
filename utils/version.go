package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cdot65/pan-os-cdss-certificate-registration/logger"
)

// Version represents a PAN-OS version
type Version struct {
	Major       int
	Feature     int
	Maintenance int
	Hotfix      int
}

// MinimumPatchedVersion represents the minimum patched version for a specific release
type MinimumPatchedVersion struct {
	Maintenance int
	Hotfix      int
}

// ParseVersion parses a version string into a Version struct
func ParseVersion(version string) (*Version, error) {
	parts := strings.Split(version, ".")
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid version format: %s", version)
	}

	// Initialize a Version struct
	v := &Version{}
	var err error

	// Parse the major version part
	v.Major, err = strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid major version: %s", parts[0])
	}

	// Parse the feature version part
	v.Feature, err = strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid feature version: %s", parts[1])
	}

	// Split the maintenance and hotfix parts
	maintenanceHotfix := strings.Split(parts[2], "-h")

	// Parse the maintenance version part
	v.Maintenance, err = strconv.Atoi(maintenanceHotfix[0])
	if err != nil {
		return nil, fmt.Errorf("invalid maintenance version: %s", maintenanceHotfix[0])
	}

	// If there's a hotfix part, parse it as well
	if len(maintenanceHotfix) > 1 {
		v.Hotfix, err = strconv.Atoi(maintenanceHotfix[1])
		if err != nil {
			return nil, fmt.Errorf("invalid hotfix version: %s", maintenanceHotfix[1])
		}
	}

	// Return the parsed Version struct
	return v, nil
}

// IsLessThan compares two Version structs
func (v *Version) IsLessThan(other *Version) bool {
	if v.Major != other.Major {
		return v.Major < other.Major
	}
	if v.Feature != other.Feature {
		return v.Feature < other.Feature
	}
	if v.Maintenance != other.Maintenance {
		return v.Maintenance < other.Maintenance
	}
	return v.Hotfix < other.Hotfix
}

// MinimumPatchedVersions represents the minimum patched versions for each PAN-OS feature release
var MinimumPatchedVersions = map[string][]MinimumPatchedVersion{
	"8.1": {
		{Maintenance: 21, Hotfix: 3},
		{Maintenance: 25, Hotfix: 3},
		{Maintenance: 26, Hotfix: 0},
	},
	"9.0": {
		{Maintenance: 16, Hotfix: 7},
		{Maintenance: 17, Hotfix: 5},
	},
	"9.1": {
		{Maintenance: 11, Hotfix: 5},
		{Maintenance: 12, Hotfix: 7},
		{Maintenance: 13, Hotfix: 5},
		{Maintenance: 14, Hotfix: 8},
		{Maintenance: 16, Hotfix: 5},
		{Maintenance: 17, Hotfix: 0},
	},
	"10.0": {
		{Maintenance: 8, Hotfix: 8},
		{Maintenance: 11, Hotfix: 4},
		{Maintenance: 12, Hotfix: 5},
	},
	"10.1": {
		{Maintenance: 3, Hotfix: 3},
		{Maintenance: 4, Hotfix: 6},
		{Maintenance: 5, Hotfix: 4},
		{Maintenance: 6, Hotfix: 8},
		{Maintenance: 7, Hotfix: 1},
		{Maintenance: 8, Hotfix: 7},
		{Maintenance: 9, Hotfix: 8},
		{Maintenance: 10, Hotfix: 5},
		{Maintenance: 11, Hotfix: 5},
		{Maintenance: 12, Hotfix: 0},
	},
	"10.2": {
		{Maintenance: 0, Hotfix: 2},
		{Maintenance: 1, Hotfix: 1},
		{Maintenance: 2, Hotfix: 4},
		{Maintenance: 3, Hotfix: 12},
		{Maintenance: 4, Hotfix: 10},
		{Maintenance: 6, Hotfix: 1},
		{Maintenance: 7, Hotfix: 3},
		{Maintenance: 8, Hotfix: 0},
	},
	"10.2-gp": {
		{Maintenance: 0, Hotfix: 3},
		{Maintenance: 1, Hotfix: 2},
		{Maintenance: 2, Hotfix: 5},
		{Maintenance: 3, Hotfix: 13},
		{Maintenance: 4, Hotfix: 16},
		{Maintenance: 5, Hotfix: 6},
		{Maintenance: 6, Hotfix: 3},
		{Maintenance: 7, Hotfix: 8},
		{Maintenance: 8, Hotfix: 3},
		{Maintenance: 9, Hotfix: 1},
	},
	"11.0": {
		{Maintenance: 0, Hotfix: 2},
		{Maintenance: 1, Hotfix: 3},
		{Maintenance: 2, Hotfix: 3},
		{Maintenance: 3, Hotfix: 3},
		{Maintenance: 4, Hotfix: 0},
	},
	"11.0-gp": {
		{Maintenance: 0, Hotfix: 3},
		{Maintenance: 1, Hotfix: 4},
		{Maintenance: 2, Hotfix: 4},
		{Maintenance: 3, Hotfix: 10},
		{Maintenance: 4, Hotfix: 1},
	},
	"11.1": {
		{Maintenance: 0, Hotfix: 2},
		{Maintenance: 1, Hotfix: 0},
	},
	"11.1-gp": {
		{Maintenance: 0, Hotfix: 3},
		{Maintenance: 1, Hotfix: 1},
		{Maintenance: 2, Hotfix: 3},
	},
}

// IsAffectedVersion checks if a given version is affected (needs to be patched)
func IsAffectedVersion(device map[string]string, isGlobalProtect bool) (bool, error) {
	major, _ := strconv.Atoi(device["parsed_version_major"])
	feature, _ := strconv.Atoi(device["parsed_version_feature"])
	maintenance, _ := strconv.Atoi(device["parsed_version_maintenance"])
	hotfix, _ := strconv.Atoi(device["parsed_version_hotfix"])

	// Check if the version is 11.2 or later
	if major > 11 || (major == 11 && feature >= 2) {
		return false, nil // Versions 11.2 and later are not affected
	}

	v := &Version{
		Major:       major,
		Feature:     feature,
		Maintenance: maintenance,
		Hotfix:      hotfix,
	}

	featureRelease := fmt.Sprintf("%d.%d", v.Major, v.Feature)
	if isGlobalProtect && (featureRelease == "10.2" || featureRelease == "11.0" || featureRelease == "11.1") {
		featureRelease += "-gp"
	}

	minVersions, ok := MinimumPatchedVersions[featureRelease]
	if !ok {
		return false, fmt.Errorf("unknown feature release: %s", featureRelease)
	}

	for _, minVersion := range minVersions {
		if v.Maintenance == minVersion.Maintenance {
			return v.Hotfix < minVersion.Hotfix, nil
		}
		if v.Maintenance < minVersion.Maintenance {
			return true, nil
		}
	}

	return false, nil
}

// FilterAffectedDevices filters the device list to only include affected devices
func FilterAffectedDevices(deviceList []map[string]string) ([]map[string]string, error) {
	var affectedDevices []map[string]string

	for _, device := range deviceList {
		isAffected, err := IsAffectedVersion(device, false) // Assuming no Global Protect for now
		if err != nil {
			return nil, fmt.Errorf("error checking device %s: %v", device["hostname"], err)
		}
		if isAffected {
			affectedDevices = append(affectedDevices, device)
		}
	}

	return affectedDevices, nil
}

// PrintDeviceList prints a formatted list of devices to the console.
// This function takes a slice of device maps and a logger, then iterates through
// the list to print each device's details in a structured format.
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
// This function reads results from a channel, prints them, and keeps track of successful
// and failed registrations. It handles timeouts and unexpected channel closures.
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
