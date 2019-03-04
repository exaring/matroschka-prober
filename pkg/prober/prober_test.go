package prober

import (
	"fmt"
	"sync"
	"testing"

	"github.com/exaring/matroschka-prober/pkg/config"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/ipv4"
)

type mockSock struct {
}

func newMockSock() *mockSock {
	return &mockSock{}
}

func (m *mockSock) WriteTo(*ipv4.Header, []byte, *ipv4.ControlMessage) error {
	return nil
}

func (m *mockSock) Read([]byte) (int, error) {
	return 0, nil
}

func (m *mockSock) Close() error {
	return nil
}

func TestProberReceive(t *testing.T) {
	tests := []struct {
		name   string
		config *config.Config
	}{
		{},
	}
}

func TestProber(t *testing.T) {
	tests := []struct {
		name   string
		config *config.Config
	}{
		{
			name: "Test #1",
			config: &config.Config{
				Routers: []config.Router{
					{
						Name:     "some01.metro01",
						DstRange: "169.254.0.0/31",
					},
				},
				Paths: []config.Path{
					{
						Name: "some01.metro01",
						Hops: []string{
							"some01.metro01",
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		test.config.ApplyDefaults()
		p := New(test.config, test.config.Paths[0], 0x00)

		mc := newMockSock()
		p.rawConn = mc
		p.udpConn = mc
		go p.sender()
		go p.receiver()

		promCh := make(chan prometheus.Metric)
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			p.Collect(promCh)
			wg.Done()
		}()

		go func() {
			x := <-promCh
			fmt.Printf("x: %v\n", x)
		}()

		wg.Wait()

		t.Errorf("Foo")
	}
}
