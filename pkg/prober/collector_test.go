package prober

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func uint64ptr(v uint64) *uint64 {
	return &v
}

type mockClock struct {
	t time.Time
}

func (m mockClock) Now() time.Time {
	return m.t
}

func TestLastFinishedMeasurement(t *testing.T) {
	tests := []struct {
		name     string
		p        *Prober
		expected int64
	}{
		{
			name: "Test #1",
			p: &Prober{
				clock: mockClock{
					t: time.Unix(1542556558, 0),
				},
				cfg: Config{
					MeasurementLengthMS: 1000,
					TimeoutMS:           200,
				},
			},
			expected: 1542556556000000000,
		},
		{
			name: "Test #2",
			p: &Prober{
				clock: mockClock{
					t: time.Unix(1542556558, 250000000),
				},
				cfg: Config{
					MeasurementLengthMS: 1000,
					TimeoutMS:           200,
				},
			},
			expected: 1542556557000000000,
		},
	}

	for _, test := range tests {
		ts := test.p.lastFinishedMeasurement()
		assert.Equalf(t, test.expected, ts, test.name)
	}
}
