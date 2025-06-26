package prober

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigEqual(t *testing.T) {
	tests := []struct {
		name     string
		c1       *Config
		c2       *Config
		expected bool
	}{
		{
			name: "identical configurations",
			c1: &Config{
				Name:              "test",
				BasePort:          8080,
				ConfiguredSrcAddr: net.IP{192, 168, 1, 1},
				SrcAddrs:          []net.IP{{192, 168, 1, 2}, {192, 168, 1, 3}},
				Hops:              []Hop{{Name: "hop1"}, {Name: "hop2"}},
				StaticLabels:      []Label{{Key: "key1", Value: "value1"}},
				TOS: TOS{
					Name:  "BE",
					Value: 0x01,
				},
				PPS:                 100,
				PayloadSizeBytes:    512,
				MeasurementLengthMS: 1000,
				TimeoutMS:           500,
				IPVersion:           4,
			},
			c2: &Config{
				Name:              "test",
				BasePort:          8080,
				ConfiguredSrcAddr: net.IP{192, 168, 1, 1},
				SrcAddrs:          []net.IP{{192, 168, 1, 2}, {192, 168, 1, 3}},
				Hops:              []Hop{{Name: "hop1"}, {Name: "hop2"}},
				StaticLabels:      []Label{{Key: "key1", Value: "value1"}},
				TOS: TOS{
					Name:  "BE",
					Value: 0x01,
				},
				PPS:                 100,
				PayloadSizeBytes:    512,
				MeasurementLengthMS: 1000,
				TimeoutMS:           500,
				IPVersion:           4,
			},
			expected: true,
		},
		{
			name: "different ConfiguredSrcAddr",
			c1: &Config{
				Name:              "test",
				ConfiguredSrcAddr: net.IP{192, 168, 1, 1},
			},
			c2: &Config{
				Name:              "test",
				ConfiguredSrcAddr: net.IP{192, 168, 1, 2},
			},
			expected: false,
		},
		{
			name: "different MeasurementLengthMS",
			c1: &Config{
				MeasurementLengthMS: 1000,
			},
			c2: &Config{
				MeasurementLengthMS: 2000,
			},
			expected: false,
		},
		{
			name: "one configuration is nil",
			c1:   nil,
			c2: &Config{
				Name: "test",
			},
			expected: false,
		},
		{
			name:     "both configurations are nil",
			c1:       nil,
			c2:       nil,
			expected: true,
		},
		{
			name: "different Hops",
			c1: &Config{
				Hops: []Hop{{Name: "hop1"}},
			},
			c2: &Config{
				Hops: []Hop{{Name: "hop2"}},
			},
			expected: false,
		},
		{
			name: "different dst Hop ranges",
			c1: &Config{
				Hops: []Hop{
					{
						Name: "hop1",
						DstRange: []net.IP{
							net.IPv4(1, 2, 3, 4),
							net.IPv4(1, 2, 3, 10),
						},
						SrcRange: nil,
					},
				},
			},
			c2: &Config{
				Hops: []Hop{
					{
						Name: "hop1",
						DstRange: []net.IP{
							net.IPv4(1, 2, 3, 4),
							net.IPv4(1, 2, 3, 11),
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "different src Hop ranges",
			c1: &Config{
				Hops: []Hop{
					{
						Name: "hop1",
						SrcRange: []net.IP{
							net.IPv4(1, 2, 3, 4),
							net.IPv4(1, 2, 3, 10),
						},
					},
				},
			},
			c2: &Config{
				Hops: []Hop{
					{
						Name: "hop1",
						SrcRange: []net.IP{
							net.IPv4(1, 2, 3, 4),
							net.IPv4(1, 2, 3, 11),
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "different StaticLabels",
			c1: &Config{
				StaticLabels: []Label{{Key: "key1", Value: "value1"}},
			},
			c2: &Config{
				StaticLabels: []Label{{Key: "key2", Value: "value2"}},
			},
			expected: false,
		},
		{
			name: "empty slices",
			c1: &Config{
				SrcAddrs:     []net.IP{},
				Hops:         []Hop{},
				StaticLabels: []Label{},
			},
			c2: &Config{
				SrcAddrs:     []net.IP{},
				Hops:         []Hop{},
				StaticLabels: []Label{},
			},
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.c1.Equal(test.c2)
			assert.Equal(t, test.expected, result)
		})
	}
}
