package prober

import (
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetIPVersion(t *testing.T) {
	tests := []struct {
		name      string
		input     *Prober
		expected  int8
		wantError bool
	}{
		{
			name: "Test version from Hop, ipv4",
			input: &Prober{
				cfg: Config{
					Hops: []Hop{
						{
							Name: "first-Hop",
							DstRange: []net.IP{
								{10, 255, 0, 0},
							},
							SrcRange: []net.IP{
								{10, 255, 1, 1},
							},
						},
					},
				},
			},
			expected:  4,
			wantError: false,
		},
		{
			name: "Test version from Hop, ipv6",
			input: &Prober{
				cfg: Config{
					Hops: []Hop{
						{
							Name: "first-Hop",
							DstRange: []net.IP{
								{0x20, 0x1, 0xD, 0xB8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
							},
							SrcRange: []net.IP{
								{0x20, 0x1, 0xD, 0xB8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
							},
						},
					},
				},
			},
			expected:  6,
			wantError: false,
		},
		{
			name: "Malformed address",
			input: &Prober{
				cfg: Config{
					Hops: []Hop{
						{
							Name: "first-Hop",
							DstRange: []net.IP{
								{10, 255},
							},
							SrcRange: []net.IP{
								{10, 255, 1},
							},
						},
					},
				},
			},
			expected:  4,
			wantError: true,
		},
	}
	for _, test := range tests {
		version, err := test.input.getIPVersion()

		if !test.wantError {
			assert.Equal(t, test.expected, version, test.name)
			assert.Equal(t, err, nil, test.name)
		}
		if test.wantError {
			expectedErr := fmt.Errorf("Couldn't determine the protocol version for address 10.255.1.")
			assert.Equal(t, err, expectedErr, test.name)
		}
	}
}

func TestCraftPacket(t *testing.T) {
	tests := []struct {
		name     string
		prober   *Prober
		pr       *probe
		expected []byte
	}{
		{
			name: "Test packet crafter",
			prober: &Prober{
				cfg: Config{
					Hops: []Hop{
						{
							Name: "first-hop",
							DstRange: []net.IP{
								{10, 255, 0, 1},
							},
							SrcRange: []net.IP{
								{10, 255, 1, 1},
							},
						},
					},
					IPProtocol: 4,
				},
				localAddr:  net.IP{10, 255, 3, 1},
				dstUDPPort: 9090,
			},
			pr: &probe{
				TimeStamp:      0,
				SequenceNumber: 5,
			},
			expected: []byte{0x0, 0x0, 0x8, 0x0, 0x45, 0x0, 0x0, 0x2c, 0x0, 0x0, 0x0, 0x0, 0x40, 0x11, 0x60, 0xc2, 0xa, 0xff, 0x1, 0x1, 0xa, 0xff, 0x3, 0x1, 0x23, 0x82, 0x23, 0x82, 0x0, 0x18, 0x9e, 0xb5, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x5, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name: "Test packet crafter ipv6",
			prober: &Prober{
				cfg: Config{
					Hops: []Hop{
						{
							Name: "first-hop",
							DstRange: []net.IP{
								{0x20, 0x1, 0xD, 0xB8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 5},
							},
							SrcRange: []net.IP{
								{0x20, 0x1, 0xD, 0xB8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
							},
						},
					},
					IPProtocol: 6,
				},
				localAddr:  net.IP{0x20, 0x1, 0xD, 0xB8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				dstUDPPort: 9090,
			},
			pr: &probe{
				TimeStamp:      0,
				SequenceNumber: 8,
			},
			expected: []byte{0x0, 0x0, 0x86, 0xdd, 0x60, 0x0, 0x0, 0x0, 0x0, 0x18, 0x11, 0x40, 0x20, 0x1, 0xd, 0xb8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x20, 0x1, 0xd, 0xb8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x23, 0x82, 0x23, 0x82, 0x0, 0x18, 0x5d, 0x40, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
	}
	for _, test := range tests {
		packet, _ := test.prober.craftPacket(test.pr)
		assert.Equal(t, test.expected, packet, test.name)
	}
}

