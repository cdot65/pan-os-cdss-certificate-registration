// Package filters utils/filters/family.go
package filters

import (
	"github.com/cdot65/pan-os-cdss-certificate-registration/config"
)

// IsAffectedFamily checks if a device's family is in the list of affected families
func IsAffectedFamily(family string, model string) bool {
	if affectedModels, ok := config.AffectedFamilies[family]; ok {
		for _, affectedModel := range affectedModels {
			if affectedModel == model {
				return true
			}
		}
	}
	return false
}

// FilterDevicesByFamily separates devices into affected and unaffected based on their family and model
func FilterDevicesByFamily(devices []map[string]string) (affected []map[string]string, unaffected []map[string]string) {
	for _, device := range devices {
		family := device["family"]
		model := device["model"]

		if IsAffectedFamily(family, model) {
			affected = append(affected, device)
		} else {
			unaffected = append(unaffected, device)
		}
	}
	return affected, unaffected
}
