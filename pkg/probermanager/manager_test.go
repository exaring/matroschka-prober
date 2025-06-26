package probermanager

import (
	"testing"

	"github.com/exaring/matroschka-prober/pkg/prober"
	"github.com/stretchr/testify/assert"
)

func TestLabels(t *testing.T) {
	tests := []struct {
		m        map[string]string
		expected []prober.Label
	}{
		{
			m: map[string]string{
				"foo": "bar",
				"abc": "def",
			},
			expected: []prober.Label{
				{
					Key:   "foo",
					Value: "bar",
				},
				{
					Key:   "abc",
					Value: "def",
				},
			},
		},
	}

	for _, test := range tests {
		res := labels(test.m)
		assert.Equal(t, test.expected, res)
	}
}
