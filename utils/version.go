package utils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cdot65/pan-os-cdss-certificate-registration/config"
)

// Version represents a PAN-OS version
type Version struct {
	Major       int
	Feature     int
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

	minVersions, ok := config.MinimumPatchedVersions[featureRelease]
	if !ok {
		// If the feature release is not in MinimumPatchedVersions
		if v.Major < 8 || (v.Major == 8 && v.Feature < 1) {
			return true, nil // Versions earlier than 8.1 are considered affected
		}
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
