package prober

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"

	"github.com/exaring/matroschka-prober/pkg/config"
	"github.com/exaring/matroschka-prober/pkg/measurement"
	"github.com/google/gopacket"
	"golang.org/x/net/ipv4"
)

const (
	mtuMax = uint16(9216)
)

// Prober keeps the state of a prober instance. There is one instance per probed path.
type Prober struct {
	dstUDPPort             uint16
	cfg                    *config.Config
	hops                   []hop
	localAddr              net.IP
	mtu                    uint16
	payload                gopacket.Payload
	probesReceived         uint64
	probesSent             uint64
	path                   config.Path
	rawConn                *ipv4.RawConn // Used to send GRE packets
	srcAddrs               []net.IP
	stop                   chan struct{}
	transitProbes          *transitProbes // Keeps track of in-flight packets
	tos                    uint8
	udpConn                *net.UDPConn // Used to receive returning packets
	measurements           *measurement.MeasurementsDB
	measurementsAggregated *measurement.MeasurementsDB
}

type hop struct {
	name     string
	dstRange []net.IP
}

func (h *hop) getAddr(s uint64) net.IP {
	return h.dstRange[s%uint64(len(h.dstRange))]
}

// New creates a new prober
func New(c *config.Config, p config.Path, tos uint8) *Prober {
	return &Prober{
		cfg:                    c,
		hops:                   confHopsToHops(c, p),
		path:                   p,
		mtu:                    mtuMax,
		transitProbes:          newTransitProbes(),
		measurements:           measurement.NewDB(),
		measurementsAggregated: measurement.NewDB(),
		srcAddrs:               generateAddrs(*p.SrcRange),
		tos:                    tos,
		payload:                make(gopacket.Payload, *p.PayloadSizeBytes),
	}
}

// Start starts the prober
func (p *Prober) Start() error {
	err := p.init()
	if err != nil {
		return fmt.Errorf("Failed to init: %v", err)
	}

	// TODO: Start service routines
	return nil
}

func confHopsToHops(cfg *config.Config, pathCfg config.Path) []hop {
	res := make([]hop, 0)

	for i := range pathCfg.Hops {
		for j := range cfg.Routers {
			if pathCfg.Hops[i] != cfg.Routers[j].Name {
				continue
			}

			h := hop{
				name:     cfg.Routers[j].Name,
				dstRange: generateAddrs(cfg.Routers[j].DstRange),
			}
			res = append(res, h)
		}
	}

	return res
}

func (p *Prober) getSrcAddr(s uint64) net.IP {
	return p.srcAddrs[s%uint64(len(p.srcAddrs))]
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

func generateAddrs(addrRange string) []net.IP {
	_, n, err := net.ParseCIDR(addrRange)
	if err != nil {
		panic(err)
	}

	baseAddr := getCIDRBase(*n)
	c := maskAddrCount(*n)
	ret := make([]net.IP, c)

	for i := uint32(0); i < c; i++ {
		ret[i] = net.IP(uint32Byte(baseAddr + i%c))
	}

	return ret
}

func getCIDRBase(n net.IPNet) uint32 {
	return uint32b(n.IP)
}

func uint32b(data []byte) (ret uint32) {
	buf := bytes.NewBuffer(data)
	binary.Read(buf, binary.BigEndian, &ret)
	return
}
