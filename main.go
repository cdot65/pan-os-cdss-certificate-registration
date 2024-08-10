package main

import (
	"flag"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/cdot65/pan-os-cdss-certificate-registration/config"
	"github.com/cdot65/pan-os-cdss-certificate-registration/devices"
	"github.com/cdot65/pan-os-cdss-certificate-registration/logger"
	"github.com/cdot65/pan-os-cdss-certificate-registration/wildfire"
)

// Main function to register WildFire on multiple devices concurrently.

// This function parses command-line flags, loads configuration, retrieves a list of devices,
// and concurrently registers WildFire on each device. It uses goroutines for parallel processing
// and reports the results for each device.

// Error:
//   Various errors may be logged and cause program termination, including:
//   - Configuration loading failures
//   - Device list retrieval failures
//   - WildFire registration failures for individual devices

// Note: This function doesn't return any value but prints results to stdout.
func main() {
	// Parse command-line flags
	cfg := parseFlags()

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

	// Print device list
	printDeviceList(deviceList, l)

	// Register WildFire
	results := make(chan string, len(deviceList))
	var wg sync.WaitGroup

	for i, device := range deviceList {
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
	printResults(results, len(deviceList), l)
}

// parseFlags parses command-line flags and returns a configuration object.

// This function sets up and parses command-line flags for various configuration options,
// including debug level, concurrency, file paths, and operational modes.

// Return:
//
//	cfg (*config.Flags): A pointer to a config.Flags struct containing parsed flag values.
func parseFlags() *config.Flags {
	cfg := &config.Flags{}
	flag.IntVar(&cfg.DebugLevel, "debug", 0, "Debug level: 0=INFO, 1=DEBUG")
	flag.IntVar(&cfg.Concurrency, "concurrency", runtime.NumCPU(), "Number of concurrent operations")
	flag.StringVar(&cfg.ConfigFile, "config", "panorama.yaml", "Path to the Panorama configuration file")
	flag.StringVar(&cfg.SecretsFile, "secrets", ".secrets.yaml", "Path to the secrets file")
	flag.StringVar(&cfg.HostnameFilter, "filter", "", "Comma-separated list of hostname patterns to filter devices")
	flag.BoolVar(&cfg.Verbose, "verbose", false, "Enable verbose logging")
	flag.BoolVar(&cfg.NoPanorama, "nopanorama", false, "Use inventory.yaml instead of querying Panorama")
	flag.Parse()
	return cfg
}

// printDeviceList prints a formatted list of devices to the console.

// This function takes a slice of device maps and a logger, then iterates through
// the list to print each device's details in a structured format.

// Attributes:
//   deviceList ([]map[string]string): A slice of maps containing device information.
//   l (*logger.Logger): A pointer to a logger for logging information.

// Error:
//
//   None

// Return:
//
//	None
func printDeviceList(deviceList []map[string]string, l *logger.Logger) {
	l.Info("Printing device list")
	fmt.Println("Device List:")
	for i, device := range deviceList {
		fmt.Printf("Device %d:\n", i+1)
		for key, value := range device {
			fmt.Printf("  %s: %s\n", key, value)
		}
		fmt.Println()
	}
}

// printResults processes and displays WildFire registration results for multiple devices.

// This function reads results from a channel, prints them, and keeps track of successful
// and failed registrations. It handles timeouts and unexpected channel closures.

// Attributes:
//   results (chan string): Channel containing registration result messages.
//   totalDevices (int): Total number of devices to process results for.
//   l (*logger.Logger): Logger instance for logging information and errors.

// Error:
//
//   None explicitly thrown, but logs unexpected channel closure and timeouts.

// Return:
//
//	None
func printResults(results chan string, totalDevices int, l *logger.Logger) {
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
