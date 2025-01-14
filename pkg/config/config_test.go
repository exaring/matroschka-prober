package config

import (
	"net"
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigApplyDefaults(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *Config
		expected *Config
	}{
		{
			name: "Test #1: loading of default settings",
			cfg:  &Config{},
			expected: &Config{
				MetricsPath:      &dfltMetricsPath,
				ListenAddressStr: &dfltListenAddress,
				BasePort:         &dfltBasePort,
				SrcRangeStr:      &dfltSrcRange,
				SrcRange: &net.IPNet{
					IP:   net.IP{169, 254, 0, 0},
					Mask: net.IPMask{255, 255, 0, 0},
				},
				Defaults: &Defaults{
					MeasurementLengthMS: &dfltMeasurementLengthMS,
					PayloadSizeBytes:    &dfltPayloadSizeBytes,
					PPS:                 &dfltPPS,
					SrcRangeStr:         &dfltSrcRange,
					SrcRange: &net.IPNet{
						IP:   net.IP{169, 254, 0, 0},
						Mask: net.IPMask{255, 255, 0, 0},
					},
					TimeoutMS: &dfltTimeoutMS,
				},
				Classes: []Class{
					{
						Name: "BE",
						TOS:  0x00,
					},
				},
			},
		},
		{
			name: "Test #2: loading default settings for paths",
			cfg: &Config{
				Paths: []Path{
					{
						Name: "Some path test",
						Hops: []string{
							"SomeRouter02.SomeMetro01",
						},
					},
				},
				Routers: []Router{
					{
						Name:        "SomeRouter02.SomeMetro01",
						SrcRangeStr: "192.168.100.0/24",
						DstRange:    parseNetwork("192.168.0.0/24"),
						SrcRange:    parseNetwork("192.168.100.0/24"),
					},
				},
			},
			expected: &Config{
				MetricsPath:      &dfltMetricsPath,
				ListenAddressStr: &dfltListenAddress,
				BasePort:         &dfltBasePort,
				SrcRangeStr:      &dfltSrcRange,
				SrcRange: &net.IPNet{
					IP:   net.IP{169, 254, 0, 0},
					Mask: net.IPMask{255, 255, 0, 0},
				},
				Defaults: &Defaults{
					MeasurementLengthMS: &dfltMeasurementLengthMS,
					PayloadSizeBytes:    &dfltPayloadSizeBytes,
					PPS:                 &dfltPPS,
					SrcRangeStr:         &dfltSrcRange,
					SrcRange: &net.IPNet{
						IP:   net.IP{169, 254, 0, 0},
						Mask: net.IPMask{255, 255, 0, 0},
					},
					TimeoutMS: &dfltTimeoutMS,
				},
				Paths: []Path{
					{
						Name: "Some path test",
						Hops: []string{
							"SomeRouter02.SomeMetro01",
						},
						MeasurementLengthMS: &dfltMeasurementLengthMS,
						PayloadSizeBytes:    &dfltPayloadSizeBytes,
						PPS:                 &dfltPPS,
						TimeoutMS:           &dfltTimeoutMS,
					},
				},
				Routers: []Router{
					{
						Name:        "SomeRouter02.SomeMetro01",
						DstRange:    parseNetwork("192.168.0.0/24"),
						SrcRange:    parseNetwork("192.168.100.0/24"),
						SrcRangeStr: "192.168.100.0/24",
					},
				},
				Classes: []Class{
					{
						Name: "BE",
						TOS:  0x00,
					},
				},
			},
		},
	}

	for _, test := range tests {
		test.cfg.ApplyDefaults()
		assert.Equal(t, test.expected, test.cfg, test.name)
	}
}

func TestGenerateAddrs(t *testing.T) {
	tests := []struct {
		addrRange   *net.IPNet
		expected    []net.IP
		shouldPanic bool
	}{
		{
			addrRange:   parseNetwork("192.168.1.0/30"),
			expected:    []net.IP{net.ParseIP("192.168.1.0"), net.ParseIP("192.168.1.1"), net.ParseIP("192.168.1.2"), net.ParseIP("192.168.1.3")},
			shouldPanic: false,
		},
		{
			addrRange:   parseNetwork("192.168.1.0/31"),
			expected:    []net.IP{net.ParseIP("192.168.1.0"), net.ParseIP("192.168.1.1")},
			shouldPanic: false,
		},
		{
			addrRange:   parseNetwork("192.168.1.0/32"),
			expected:    []net.IP{net.ParseIP("192.168.1.0")},
			shouldPanic: false,
		},
		{
			addrRange:   parseNetwork("2001:db8::/126"),
			expected:    []net.IP{net.ParseIP("2001:db8::"), net.ParseIP("2001:db8::1"), net.ParseIP("2001:db8::2"), net.ParseIP("2001:db8::3")},
			shouldPanic: false,
		},
		{
			addrRange:   parseNetwork("invalid-range"),
			expected:    nil,
			shouldPanic: true,
		},
		{
			addrRange:   parseNetwork("2001:db8::/64"),
			expected:    nil,
			shouldPanic: true,
		},
	}

	for _, tt := range tests {
		if tt.shouldPanic {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("Expected panic for addrRange %s, but did not panic", tt.addrRange)
				}
			}()
		}

		result := GenerateAddrs(tt.addrRange)
		if !tt.shouldPanic && !equal(result, tt.expected) {
			t.Errorf("GenerateAddrs(%s) = %v; want %v", tt.addrRange, result, tt.expected)
		}
	}
}

func equal(a, b []net.IP) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !a[i].Equal(b[i]) {
			return false
		}
	}
	return true
}

func parseNetwork(network string) *net.IPNet {
	_, ret, _ := net.ParseCIDR(network)
	return ret
}
