package filters

import (
	"reflect"
	"testing"
)

func TestIsAffectedFamily(t *testing.T) {
	tests := []struct {
		name     string
		family   string
		model    string
		expected bool
	}{
		{"Affected PA-220", "220", "PA-220", true},
		{"Unaffected PA-460", "400", "PA-460", false},
		{"Affected PA-850", "800", "PA-850", true},
		{"Unaffected PA-5450", "5400", "PA-5450", false},
		{"Non-existent family", "1000", "PA-1000", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsAffectedFamily(tt.family, tt.model)
			if result != tt.expected {
				t.Errorf("IsAffectedFamily(%q, %q) = %v, want %v", tt.family, tt.model, result, tt.expected)
			}
		})
	}
}

func TestFilterDevicesByFamily(t *testing.T) {
	devices := []map[string]string{
		{"family": "220", "model": "PA-220"},
		{"family": "400", "model": "PA-460"},
		{"family": "800", "model": "PA-850"},
		{"family": "5400", "model": "PA-5450"},
		{"family": "vm", "model": "PA-VM"},
	}

	expectedAffected := []map[string]string{
		{"family": "220", "model": "PA-220"},
		{"family": "800", "model": "PA-850"},
		{"family": "vm", "model": "PA-VM"},
	}

	expectedUnaffected := []map[string]string{
		{"family": "400", "model": "PA-460"},
		{"family": "5400", "model": "PA-5450"},
	}

	affected, unaffected := FilterDevicesByFamily(devices)

	if !reflect.DeepEqual(affected, expectedAffected) {
		t.Errorf("Affected devices mismatch.\nGot: %v\nWant: %v", affected, expectedAffected)
	}

	if !reflect.DeepEqual(unaffected, expectedUnaffected) {
		t.Errorf("Unaffected devices mismatch.\nGot: %v\nWant: %v", unaffected, expectedUnaffected)
	}
}
