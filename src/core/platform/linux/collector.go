package linux

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/minibeast/usb-agent/src/core/platform/types"
)

// Collector implements platform.Collector for Linux systems
type Collector struct{}

// NewCollector creates a new Linux collector
// Complexity: O(1)
func NewCollector() (*Collector, error) {
	return &Collector{}, nil
}

// GetSystemInfo retrieves Linux system information
// Complexity: O(1) - reads fixed-size system files
func (c *Collector) GetSystemInfo(ctx context.Context) (*types.SystemInfo, error) {
	info := &types.SystemInfo{
		OSName: "Linux",
	}

	// Get hostname
	if hostname, err := os.Hostname(); err == nil {
		info.Hostname = hostname
	} else {
		info.Hostname = "unknown"
	}

	// Get OS version from /etc/os-release
	if version, err := c.getOSVersion(); err == nil {
		info.OSVersion = version
	} else {
		info.OSVersion = "unknown"
	}

	// Get kernel version
	if build, err := c.getKernelVersion(); err == nil {
		info.OSBuild = build
	} else {
		info.OSBuild = "unknown"
	}

	// Get timezone
	if tz, err := c.getTimezone(); err == nil {
		info.Timezone = tz
	} else {
		info.Timezone = "UTC"
	}

	return info, nil
}

// GetNetworkInfo retrieves Linux network configuration
// Complexity: O(n) where n = number of network interfaces
func (c *Collector) GetNetworkInfo(ctx context.Context) (*types.NetworkInfo, error) {
	info := &types.NetworkInfo{
		Interfaces: []types.NetworkInterface{},
		WiFiSSIDs:  []string{},
	}

	// Parse /proc/net/dev for interfaces
	interfaces, err := c.getNetworkInterfaces()
	if err == nil {
		info.Interfaces = interfaces
	}

	// Get WiFi SSIDs (best-effort)
	ssids, err := c.getWiFiSSIDs()
	if err == nil {
		info.WiFiSSIDs = ssids
	}

	// Sort for determinism
	sort.Slice(info.Interfaces, func(i, j int) bool {
		return info.Interfaces[i].Name < info.Interfaces[j].Name
	})
	sort.Strings(info.WiFiSSIDs)

	return info, nil
}

// GetHardwareInfo retrieves Linux hardware identifiers
// Complexity: O(1) - reads fixed system files
func (c *Collector) GetHardwareInfo(ctx context.Context) (*types.HardwareInfo, error) {
	info := &types.HardwareInfo{
		SerialNumber: "unknown",
		HardwareUUID: "unknown",
	}

	// Try to read machine-id (best hardware identifier on Linux)
	if uuid, err := c.getMachineID(); err == nil {
		info.HardwareUUID = uuid
	}

	// Try DMI product serial (requires root, graceful degradation)
	if serial, err := c.getDMISerial(); err == nil {
		info.SerialNumber = serial
	}

	return info, nil
}

// GetPIIInfo retrieves Linux user information
// Complexity: O(u) where u = number of users
func (c *Collector) GetPIIInfo(ctx context.Context) (*types.PIIInfo, error) {
	info := &types.PIIInfo{
		Users:          []types.User{},
		LoggedInUsers:  []string{},
		HomeDirs:       []string{},
		RecentProfiles: []types.UserProfile{},
		PrimaryEmail:   "unknown",
	}

	// Get all local users from /etc/passwd
	users, err := c.getLocalUsers()
	if err == nil {
		info.Users = users
		for _, u := range users {
			if u.Username != "" && !strings.HasPrefix(u.Username, "_") {
				info.HomeDirs = append(info.HomeDirs, "/home/"+u.Username)
			}
		}
	}

	// Get currently logged-in users
	loggedIn, err := c.getLoggedInUsers()
	if err == nil {
		info.LoggedInUsers = loggedIn
	}

	// Sort for determinism
	sort.Slice(info.Users, func(i, j int) bool {
		return info.Users[i].Username < info.Users[j].Username
	})
	sort.Strings(info.LoggedInUsers)
	sort.Strings(info.HomeDirs)

	return info, nil
}

// Helper functions

func (c *Collector) getOSVersion() (string, error) {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "VERSION_ID=") {
			version := strings.TrimPrefix(line, "VERSION_ID=")
			version = strings.Trim(version, "\"")
			return version, nil
		}
	}
	return "unknown", nil
}

func (c *Collector) getKernelVersion() (string, error) {
	data, err := os.ReadFile("/proc/version")
	if err != nil {
		return "", err
	}

	fields := strings.Fields(string(data))
	if len(fields) >= 3 {
		return fields[2], nil
	}
	return "unknown", nil
}

func (c *Collector) getTimezone() (string, error) {
	// Read /etc/timezone
	data, err := os.ReadFile("/etc/timezone")
	if err == nil {
		tz := strings.TrimSpace(string(data))
		if tz != "" {
			return tz, nil
		}
	}

	// Fallback: check TZ environment variable
	if tz := os.Getenv("TZ"); tz != "" {
		return tz, nil
	}

	// Fallback: UTC
	return time.Local.String(), nil
}

func (c *Collector) getNetworkInterfaces() ([]types.NetworkInterface, error) {
	interfaces := []types.NetworkInterface{}

	// Read /sys/class/net for interface names
	entries, err := os.ReadDir("/sys/class/net")
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		name := entry.Name()
		if name == "lo" {
			continue // Skip loopback
		}

		iface := types.NetworkInterface{
			Name:       name,
			IPAddress:  "unknown",
			MACAddress: "unknown",
		}

		// Read MAC address
		macPath := filepath.Join("/sys/class/net", name, "address")
		if data, err := os.ReadFile(macPath); err == nil {
			iface.MACAddress = strings.TrimSpace(string(data))
		}

		// Get IP address using ip command (best-effort)
		if ip, err := c.getInterfaceIP(name); err == nil {
			iface.IPAddress = ip
		}

		interfaces = append(interfaces, iface)
	}

	return interfaces, nil
}

func (c *Collector) getInterfaceIP(ifaceName string) (string, error) {
	cmd := exec.Command("ip", "addr", "show", ifaceName)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "inet ") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				// Remove CIDR notation
				ip := strings.Split(fields[1], "/")[0]
				return ip, nil
			}
		}
	}
	return "", fmt.Errorf("no IP found")
}

func (c *Collector) getWiFiSSIDs() ([]string, error) {
	ssids := []string{}

	// Try NetworkManager (common on modern Linux)
	nmPath := "/etc/NetworkManager/system-connections"
	if entries, err := os.ReadDir(nmPath); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				ssids = append(ssids, entry.Name())
			}
		}
	}

	return ssids, nil
}

func (c *Collector) getMachineID() (string, error) {
	// Try /etc/machine-id first
	data, err := os.ReadFile("/etc/machine-id")
	if err == nil {
		return strings.TrimSpace(string(data)), nil
	}

	// Fallback: /var/lib/dbus/machine-id
	data, err = os.ReadFile("/var/lib/dbus/machine-id")
	if err == nil {
		return strings.TrimSpace(string(data)), nil
	}

	return "", err
}

func (c *Collector) getDMISerial() (string, error) {
	// Requires root access, graceful degradation
	data, err := os.ReadFile("/sys/class/dmi/id/product_serial")
	if err != nil {
		return "unknown", nil // Not an error, just no access
	}
	return strings.TrimSpace(string(data)), nil
}

func (c *Collector) getLocalUsers() ([]types.User, error) {
	users := []types.User{}

	file, err := os.Open("/etc/passwd")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, ":")
		if len(fields) >= 5 {
			username := fields[0]
			uid := fields[2]
			fullName := fields[4]

			// Filter out system users (UID < 1000)
			if username != "" && !strings.HasPrefix(username, "_") {
				users = append(users, types.User{
					Username: username,
					FullName: fullName,
					UID:      uid,
				})
			}
		}
	}

	return users, scanner.Err()
}

func (c *Collector) getLoggedInUsers() ([]string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return []string{}, nil
	}

	// Simple approach: return current user
	return []string{currentUser.Username}, nil
}
