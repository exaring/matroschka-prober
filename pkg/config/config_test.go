package config

import (
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
				Defaults: &Defaults{
					BasePort:                      &dfltBasePort,
					MeasurementLengthMS:           &dfltMeasurementLengthMS,
					MeasurementLengthAggregatedMS: &dfltMeasurementLengthAggregatedMS,
					PayloadSizeBytes:              &dfltPayloadSizeBytes,
					PPS:                           &dfltPPS,
					SpoofReplySrc:                 &dfltSpoofReplySrc,
					SrcRange:                      &dfltSrcRange,
					TimeoutMS:                     &dfltTimeoutMS,
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
					},
				},
			},
			expected: &Config{
				Defaults: &Defaults{
					BasePort:                      &dfltBasePort,
					MeasurementLengthMS:           &dfltMeasurementLengthMS,
					MeasurementLengthAggregatedMS: &dfltMeasurementLengthAggregatedMS,
					PayloadSizeBytes:              &dfltPayloadSizeBytes,
					PPS:                           &dfltPPS,
					SpoofReplySrc:                 &dfltSpoofReplySrc,
					SrcRange:                      &dfltSrcRange,
					TimeoutMS:                     &dfltTimeoutMS,
				},
				Paths: []Path{
					{
						Name: "Some path test",
						Hops: []string{
							"SomeRouter02.SomeMetro01",
						},
						BasePort:                      &dfltBasePort,
						MeasurementLengthMS:           &dfltMeasurementLengthMS,
						MeasurementLengthAggregatedMS: &dfltMeasurementLengthAggregatedMS,
						PayloadSizeBytes:              &dfltPayloadSizeBytes,
						PPS:                           &dfltPPS,
						SpoofReplySrc:                 &dfltSpoofReplySrc,
						SrcRange:                      &dfltSrcRange,
						TimeoutMS:                     &dfltTimeoutMS,
					},
				},
				Routers: []Router{
					{
						Name:     "SomeRouter02.SomeMetro01",
						DstRange: "192.168.0.0/24",
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
		test.cfg.applyDefaults()
		assert.Equal(t, test.expected, test.cfg, test.name)
	}
}
