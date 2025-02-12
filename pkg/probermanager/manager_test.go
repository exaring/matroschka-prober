package probermanager

import (
	"net"
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

func TestIpListsEqual(t *testing.T) {
	tests := []struct {
		a        []net.IP
		b        []net.IP
		expected bool
	}{
		{
			a: []net.IP{
				{
					1, 1, 1, 1,
				},
				{
					2, 2, 2, 2,
				},
			},
			b: []net.IP{
				{
					1, 1, 1, 1,
				},
				{
					2, 2, 2, 2,
				},
			},
			expected: true,
		},
		{
			a: []net.IP{
				{
					1, 1, 1, 1,
				},
				{
					2, 2, 2, 2,
				},
			},
			b: []net.IP{
				{
					2, 2, 2, 2,
				},
				{
					1, 1, 1, 1,
				},
			},
			expected: false,
		},
	}

	for _, test := range tests {
		res := ipListsEqual(test.a, test.b)
		assert.Equal(t, test.expected, res)
	}
}
