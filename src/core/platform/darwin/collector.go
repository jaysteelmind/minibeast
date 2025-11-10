package darwin

import (
	"context"
	"os"
	"os/exec"
	"os/user"
	"sort"
	"strings"
	"time"

	"github.com/minibeast/usb-agent/src/core/platform/types"
)

// Collector implements platform.Collector for macOS systems
type Collector struct{}

// NewCollector creates a new Darwin collector
// Complexity: O(1)
func NewCollector() (*Collector, error) {
	return &Collector{}, nil
}

// GetSystemInfo retrieves macOS system information
// Complexity: O(1)
func (c *Collector) GetSystemInfo(ctx context.Context) (*types.SystemInfo, error) {
	info := &types.SystemInfo{
		OSName: "Darwin",
	}

	// Get hostname
	if hostname, err := os.Hostname(); err == nil {
		info.Hostname = hostname
	} else {
		info.Hostname = "unknown"
	}

	// Get macOS version using sw_vers
	if version, err := c.getSystemVersion(); err == nil {
		info.OSVersion = version
	} else {
		info.OSVersion = "unknown"
	}

	// Get build number
	if build, err := c.getBuildVersion(); err == nil {
		info.OSBuild = build
	} else {
		info.OSBuild = "unknown"
	}

	// Get timezone
	info.Timezone = time.Local.String()

	return info, nil
}

// GetNetworkInfo retrieves macOS network configuration
// Complexity: O(n) where n = number of network interfaces
func (c *Collector) GetNetworkInfo(ctx context.Context) (*types.NetworkInfo, error) {
	info := &types.NetworkInfo{
		Interfaces: []types.NetworkInterface{},
		WiFiSSIDs:  []string{},
	}

	// Get network interfaces using ifconfig
	interfaces, err := c.getNetworkInterfaces()
	if err == nil {
		info.Interfaces = interfaces
	}

	// Get WiFi SSIDs
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

// GetHardwareInfo retrieves macOS hardware identifiers
// Complexity: O(1)
func (c *Collector) GetHardwareInfo(ctx context.Context) (*types.HardwareInfo, error) {
	info := &types.HardwareInfo{
		SerialNumber: "unknown",
		HardwareUUID: "unknown",
	}

	// Get hardware UUID using ioreg
	if uuid, err := c.getHardwareUUID(); err == nil {
		info.HardwareUUID = uuid
	}

	// Get serial number using ioreg
	if serial, err := c.getSerialNumber(); err == nil {
		info.SerialNumber = serial
	}

	return info, nil
}

// GetPIIInfo retrieves macOS user information
// Complexity: O(u) where u = number of users
func (c *Collector) GetPIIInfo(ctx context.Context) (*types.PIIInfo, error) {
	info := &types.PIIInfo{
		Users:          []types.User{},
		LoggedInUsers:  []string{},
		HomeDirs:       []string{},
		RecentProfiles: []types.UserProfile{},
		PrimaryEmail:   "unknown",
	}

	// Get local users using dscl
	users, err := c.getLocalUsers()
	if err == nil {
		info.Users = users
		for _, u := range users {
			info.HomeDirs = append(info.HomeDirs, "/Users/"+u.Username)
		}
	}

	// Get currently logged-in user
	currentUser, err := user.Current()
	if err == nil {
		info.LoggedInUsers = []string{currentUser.Username}
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

func (c *Collector) getSystemVersion() (string, error) {
	cmd := exec.Command("sw_vers", "-productVersion")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func (c *Collector) getBuildVersion() (string, error) {
	cmd := exec.Command("sw_vers", "-buildVersion")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func (c *Collector) getNetworkInterfaces() ([]types.NetworkInterface, error) {
	interfaces := []types.NetworkInterface{}

	cmd := exec.Command("ifconfig")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	var currentIface *types.NetworkInterface

	for _, line := range lines {
		// New interface starts at column 0
		if len(line) > 0 && line[0] != ' ' && line[0] != '\t' {
			if strings.Contains(line, ":") {
				parts := strings.Split(line, ":")
				name := parts[0]
				if name != "lo0" { // Skip loopback
					if currentIface != nil {
						interfaces = append(interfaces, *currentIface)
					}
					currentIface = &types.NetworkInterface{
						Name:       name,
						IPAddress:  "unknown",
						MACAddress: "unknown",
					}
				}
			}
		} else if currentIface != nil {
			// Parse interface properties
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "inet ") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					currentIface.IPAddress = fields[1]
				}
			} else if strings.HasPrefix(line, "ether ") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					currentIface.MACAddress = fields[1]
				}
			}
		}
	}

	if currentIface != nil {
		interfaces = append(interfaces, *currentIface)
	}

	return interfaces, nil
}

func (c *Collector) getWiFiSSIDs() ([]string, error) {
	ssids := []string{}

	// Get known WiFi networks using airport utility
	cmd := exec.Command("/System/Library/PrivateFrameworks/Apple80211.framework/Versions/Current/Resources/airport", "-s")
	output, err := cmd.Output()
	if err != nil {
		return ssids, nil // Best-effort, not fatal
	}

	lines := strings.Split(string(output), "\n")
	for i, line := range lines {
		if i == 0 {
			continue // Skip header
		}
		fields := strings.Fields(line)
		if len(fields) > 0 {
			ssid := fields[0]
			if ssid != "" {
				ssids = append(ssids, ssid)
			}
		}
	}

	return ssids, nil
}

func (c *Collector) getHardwareUUID() (string, error) {
	cmd := exec.Command("ioreg", "-rd1", "-c", "IOPlatformExpertDevice")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "IOPlatformUUID") {
			parts := strings.Split(line, "\"")
			if len(parts) >= 4 {
				return parts[3], nil
			}
		}
	}

	return "", nil
}

func (c *Collector) getSerialNumber() (string, error) {
	cmd := exec.Command("ioreg", "-rd1", "-c", "IOPlatformExpertDevice")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "IOPlatformSerialNumber") {
			parts := strings.Split(line, "\"")
			if len(parts) >= 4 {
				return parts[3], nil
			}
		}
	}

	return "", nil
}

func (c *Collector) getLocalUsers() ([]types.User, error) {
	users := []types.User{}

	cmd := exec.Command("dscl", ".", "-list", "/Users")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		username := strings.TrimSpace(line)
		// Filter system users
		if username != "" && !strings.HasPrefix(username, "_") && username != "daemon" && username != "nobody" {
			users = append(users, types.User{
				Username: username,
				FullName: username, // Can be enhanced with dscl query
				UID:      "",       // Can be enhanced with id command
			})
		}
	}

	return users, nil
}
