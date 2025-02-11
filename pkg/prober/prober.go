package prober

import (
	"fmt"
	"net"
	"time"

	"github.com/exaring/matroschka-prober/pkg/measurement"
	"github.com/google/gopacket"
)

const (
	mtuMax = uint16(9216)
)

// Prober keeps the state of a prober instance. There is one instance per probed path.
type Prober struct {
	cfg            Config
	dstUDPPort     uint16
	localAddr      net.IP
	clock          clock
	mtu            uint16
	payload        gopacket.Payload
	probesReceived uint64
	probesSent     uint64
	rawConn        rawSocket // Used to send GRE packets
	stop           chan struct{}
	transitProbes  *transitProbes // Keeps track of in-flight packets
	udpConn        udpSocket      // Used to receive returning packets
	measurements   *measurement.MeasurementsDB
	latePackets    uint64
}

// Config is the configuration of a prober
type Config struct {
	Name                string
	BasePort            uint16
	ConfiguredSrcAddr   net.IP
	SrcAddrs            []net.IP
	Hops                []Hop
	StaticLabels        []Label
	TOS                 TOS
	PPS                 uint64
	PayloadSizeBytes    uint64
	MeasurementLengthMS uint64
	TimeoutMS           uint64
	IPVersion           uint8
}

// TOS represents a type of service mapping
type TOS struct {
	Name  string
	Value uint8
}

// Hop represents a hop on a path to be probed
type Hop struct {
	Name     string
	DstRange []net.IP
	SrcRange []net.IP
}

func (h *Hop) getAddr(s uint64) net.IP {
	return h.DstRange[s%uint64(len(h.DstRange))]
}

// New creates a new prober
func New(c Config) *Prober {
	pr := &Prober{
		cfg:           c,
		clock:         realClock{},
		mtu:           mtuMax,
		transitProbes: newTransitProbes(),
		measurements:  measurement.NewDB(),
		stop:          make(chan struct{}),
		payload:       make(gopacket.Payload, c.PayloadSizeBytes),
	}

	return pr
}

func (p *Prober) Config() *Config {
	return &p.cfg
}

// Start starts the prober
func (p *Prober) Start() error {
	err := p.init()
	if err != nil {
		return fmt.Errorf("Failed to init: %v", err)
	}

	go p.rttTimeoutChecker()
	go p.sender()
	go p.receiver()
	go p.cleaner()
	return nil
}

// Stop stops the prober
func (p *Prober) Stop() {
	close(p.stop)
}

func (p *Prober) cleaner() {
	for {
		select {
		case <-p.stop:
			return
		default:
			time.Sleep(time.Second)
			p.measurements.RemoveOlder(p.lastFinishedMeasurement())
		}
	}
}

func (p *Prober) getSrcAddr(s uint64) net.IP {
	return p.cfg.SrcAddrs[s%uint64(len(p.cfg.SrcAddrs))]
}

func (p *Prober) init() error {
	err := p.initRawSocket()
	if err != nil {
		return fmt.Errorf("Unable to initialize RAW socket: %v", err)
	}

	err = p.initUDPSocket()
	if err != nil {
		return fmt.Errorf("Unable to initialize UDP socket: %v", err)
	}

	return nil
}
