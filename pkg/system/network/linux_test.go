package network

import (
	"testing"

	"github.com/mdlayher/wifi"
)

func TestIsScannableWifiInterface(t *testing.T) {
	tests := []struct {
		name     string
		iface    *wifi.Interface
		expected bool
	}{
		{
			name: "station interface is scannable",
			iface: &wifi.Interface{
				Name: "wlan0",
				Type: wifi.InterfaceTypeStation,
			},
			expected: true,
		},
		{
			name: "ap interface is not scannable",
			iface: &wifi.Interface{
				Name: "ap0",
				Type: wifi.InterfaceTypeAP,
			},
			expected: false,
		},
		{
			name: "monitor interface is not scannable",
			iface: &wifi.Interface{
				Name: "mon0",
				Type: wifi.InterfaceTypeMonitor,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isScannableWifiInterface(tt.iface); got != tt.expected {
				t.Fatalf("isScannableWifiInterface() = %v, want %v", got, tt.expected)
			}
		})
	}
}
