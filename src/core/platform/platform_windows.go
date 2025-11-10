//go:build windows

package platform

import "github.com/minibeast/usb-agent/src/core/platform/windows"

func newCollector() (Collector, error) {
	return windows.NewCollector()
}
