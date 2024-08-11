package utils

import (
	"reflect"
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    *Version
		wantErr bool
	}{
		{"Valid version", "10.1.6-h3", &Version{10, 1, 6, 3}, false},
		{"Valid version no hotfix", "10.1.6", &Version{10, 1, 6, 0}, false},
		{"Invalid version", "10.1", nil, true},
		{"Invalid major", "a.1.6", nil, true},
		{"Invalid feature", "10.b.6", nil, true},
		{"Invalid maintenance", "10.1.c", nil, true},
		{"Invalid hotfix", "10.1.6-hd", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseVersion(tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersion_IsLessThan(t *testing.T) {
	tests := []struct {
		name  string
		v     *Version
		other *Version
		want  bool
	}{
		{"Less major", &Version{9, 1, 6, 3}, &Version{10, 1, 6, 3}, true},
		{"Equal major, less feature", &Version{10, 0, 6, 3}, &Version{10, 1, 6, 3}, true},
		{"Equal major and feature, less maintenance", &Version{10, 1, 5, 3}, &Version{10, 1, 6, 3}, true},
		{"Equal major, feature, and maintenance, less hotfix", &Version{10, 1, 6, 2}, &Version{10, 1, 6, 3}, true},
		{"Equal versions", &Version{10, 1, 6, 3}, &Version{10, 1, 6, 3}, false},
		{"Greater version", &Version{10, 1, 6, 4}, &Version{10, 1, 6, 3}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v.IsLessThan(tt.other); got != tt.want {
				t.Errorf("Version.IsLessThan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsAffectedVersion(t *testing.T) {
	tests := []struct {
		name            string
		device          map[string]string
		isGlobalProtect bool
		want            bool
		wantErr         bool
		expectedErrMsg  string
	}{
		{
			name: "Affected version",
			device: map[string]string{
				"parsed_version_major":       "10",
				"parsed_version_feature":     "1",
				"parsed_version_maintenance": "6",
				"parsed_version_hotfix":      "2",
			},
			isGlobalProtect: false,
			want:            true,
			wantErr:         false,
		},
		{
			name: "Not affected version",
			device: map[string]string{
				"parsed_version_major":       "10",
				"parsed_version_feature":     "1",
				"parsed_version_maintenance": "6",
				"parsed_version_hotfix":      "8",
			},
			isGlobalProtect: false,
			want:            false,
			wantErr:         false,
		},
		{
			name: "GlobalProtect affected version",
			device: map[string]string{
				"parsed_version_major":       "10",
				"parsed_version_feature":     "2",
				"parsed_version_maintenance": "3",
				"parsed_version_hotfix":      "0",
			},
			isGlobalProtect: true,
			want:            true,
			wantErr:         false,
		},
		{
			name: "Version earlier than 8.1",
			device: map[string]string{
				"parsed_version_major":       "8",
				"parsed_version_feature":     "0",
				"parsed_version_maintenance": "0",
				"parsed_version_hotfix":      "0",
			},
			isGlobalProtect: false,
			want:            true,
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsAffectedVersion(tt.device, tt.isGlobalProtect)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsAffectedVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && err.Error() != tt.expectedErrMsg {
				t.Errorf("IsAffectedVersion() error = %v, expectedErrMsg %v", err, tt.expectedErrMsg)
				return
			}
			if got != tt.want {
				t.Errorf("IsAffectedVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterAffectedDevices(t *testing.T) {
	devices := []map[string]string{
		{
			"hostname":                   "device1",
			"parsed_version_major":       "10",
			"parsed_version_feature":     "1",
			"parsed_version_maintenance": "6",
			"parsed_version_hotfix":      "2",
		},
		{
			"hostname":                   "device2",
			"parsed_version_major":       "10",
			"parsed_version_feature":     "1",
			"parsed_version_maintenance": "6",
			"parsed_version_hotfix":      "8",
		},
		{
			"hostname":                   "device3",
			"parsed_version_major":       "10",
			"parsed_version_feature":     "1",
			"parsed_version_maintenance": "5",
			"parsed_version_hotfix":      "0",
		},
	}

	want := []map[string]string{devices[0], devices[2]}

	got, err := FilterAffectedDevices(devices)
	if err != nil {
		t.Errorf("FilterAffectedDevices() error = %v", err)
		return
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("FilterAffectedDevices() = %v, want %v", got, want)
	}
}
