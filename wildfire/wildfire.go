// Package wildfire/wildfire.go
package wildfire

import (
	"fmt"
	"strings"
	"time"

	"github.com/cdot65/pan-os-cdss-certificate-registration/logger"
	"github.com/scrapli/scrapligo/driver/generic"
	"github.com/scrapli/scrapligo/driver/options"
	"github.com/scrapli/scrapligo/transport"
)

// RegisterWildFire registers a device with WildFire public cloud service.

// This function connects to a specified device using SSH, sends a WildFire
// registration command, and verifies the output. It handles connection
// errors and unexpected command outputs.

// Attributes:
//   device (map[string]string): Device information including hostname and IP address.
//   username (string): SSH username for device authentication.
//   password (string): SSH password for device authentication.
//   l (*logger.Logger): Logger instance for debug output.

// Error:
//   error: Connection failures, command execution errors, or unexpected outputs.

// Return:
//   error: nil if successful, otherwise an error describing the failure.

func RegisterWildFire(device map[string]string, username, password string, l *logger.Logger) error {
	l.Debug("Attempting to connect to", device["hostname"], "at", device["ip-address"])

	d, err := generic.NewDriver(
		device["ip-address"],
		options.WithAuthNoStrictKey(),
		options.WithAuthUsername(username),
		options.WithAuthPassword(password),
		options.WithTimeoutSocket(45*time.Second),
		options.WithTimeoutOps(45*time.Second),
		options.WithTransportType(transport.StandardTransport),
		options.WithSSHConfigFile(""),
		options.WithPort(22),
	)
	if err != nil {
		l.Debug("Failed to create driver:", err)
		return fmt.Errorf("failed to create driver: %v", err)
	}

	err = d.Open()
	if err != nil {
		l.Debug("Failed to open connection:", err)
		return fmt.Errorf("failed to open connection: %v", err)
	}
	// Only defer Close() if the connection was successfully opened
	defer func() {
		if err := d.Close(); err != nil {
			l.Debug("Failed to close connection:", err)
		}
	}()

	l.Debug("Successfully connected to", device["hostname"])

	cmd := "request wildfire registration channel public"
	l.Debug("Sending WildFire registration command to", device["hostname"], "Command:", cmd)

	r, err := d.SendCommand(cmd)
	if err != nil {
		l.Debug("Failed to send command:", err)
		return fmt.Errorf("failed to send command: %v", err)
	}
	if r.Failed != nil {
		l.Debug("Command failed:", r.Failed)
		return fmt.Errorf("command failed: %v", r.Failed)
	}

	l.Debug("Command output for", device["hostname"], ":", r.Result)

	if !strings.Contains(r.Result, "WildFire registration for Public Cloud is triggered") {
		l.Debug("Unexpected command output for", device["hostname"])
		return fmt.Errorf("unexpected command output: %s", r.Result)
	}

	l.Debug("Successfully registered WildFire for", device["hostname"])
	return nil
}
