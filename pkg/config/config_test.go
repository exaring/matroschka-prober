package config

import (
	"testing"
	"net"

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
				MetricsPath:   &dfltMetricsPath,
				ListenAddress: &dfltListenAddress,
				BasePort:      &dfltBasePort,
				SrcRange:      &dfltSrcRange,
				Defaults: &Defaults{
					MeasurementLengthMS: &dfltMeasurementLengthMS,
					PayloadSizeBytes:    &dfltPayloadSizeBytes,
					PPS:                 &dfltPPS,
					SrcRange:            &dfltSrcRange,
					TimeoutMS:           &dfltTimeoutMS,
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
						Name:     "SomeRouter02.SomeMetro01",
						DstRange: "192.168.0.0/24",
						SrcRange: "192.168.100.0/24",
					},
				},
			},
			expected: &Config{
				MetricsPath:   &dfltMetricsPath,
				ListenAddress: &dfltListenAddress,
				BasePort:      &dfltBasePort,
				SrcRange:      &dfltSrcRange,
				Defaults: &Defaults{
					MeasurementLengthMS: &dfltMeasurementLengthMS,
					PayloadSizeBytes:    &dfltPayloadSizeBytes,
					PPS:                 &dfltPPS,
					SrcRange:            &dfltSrcRange,
					TimeoutMS:           &dfltTimeoutMS,
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
						Name:     "SomeRouter02.SomeMetro01",
						DstRange: "192.168.0.0/24",
						SrcRange: "192.168.100.0/24",
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
		addrRange string
		expected  []net.IP
		shouldPanic bool
	}{
		{
			addrRange: "192.168.1.0/30",
			expected:  []net.IP{net.ParseIP("192.168.1.0"), net.ParseIP("192.168.1.1"), net.ParseIP("192.168.1.2"), net.ParseIP("192.168.1.3")},
			shouldPanic: false,
		},
		{
			addrRange: "192.168.1.0/31",
			expected:  []net.IP{net.ParseIP("192.168.1.0"), net.ParseIP("192.168.1.1")},
			shouldPanic: false,
		},
		{
			addrRange: "192.168.1.0/32",
			expected:  []net.IP{net.ParseIP("192.168.1.0")},
			shouldPanic: false,
		},
		{
			addrRange: "2001:db8::/126",
			expected:  []net.IP{net.ParseIP("2001:db8::"), net.ParseIP("2001:db8::1"), net.ParseIP("2001:db8::2"), net.ParseIP("2001:db8::3")},
			shouldPanic: false,
		},
		{
			addrRange: "invalid-range",
			expected:  nil,
			shouldPanic: true,
		},
		{
			addrRange: "2001:db8::/64",
			expected: nil,
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
