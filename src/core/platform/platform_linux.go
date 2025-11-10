//go:build linux

package platform

import "github.com/minibeast/usb-agent/src/core/platform/linux"

func newCollector() (Collector, error) {
	return linux.NewCollector()
}
