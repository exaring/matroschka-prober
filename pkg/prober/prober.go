package prober

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"github.com/exaring/matroschka-prober/pkg/config"
	"github.com/exaring/matroschka-prober/pkg/measurement"
	"github.com/google/gopacket"
	"github.com/pkg/errors"
)

const (
	mtuMax = uint16(9216)
)

// Prober keeps the state of a prober instance. There is one instance per probed path.
type Prober struct {
	dstUDPPort        uint16
	cfg               *config.Config
	clock             clock
	hops              []hop
	localAddr         net.IP
	mtu               uint16
	payload           gopacket.Payload
	probesReceived    uint64
	probesSent        uint64
	path              config.Path
	rawConn           rawSocket // Used to send GRE packets
	configuredSrcAddr net.IP
	srcAddrs          []net.IP
	stop              chan struct{}
	transitProbes     *transitProbes // Keeps track of in-flight packets
	tos               uint8
	udpConn           udpSocket // Used to receive returning packets
	measurements      *measurement.MeasurementsDB
}

type hop struct {
	name     string
	dstRange []net.IP
	srcRange []net.IP
}

func (h *hop) getAddr(s uint64) net.IP {
	return h.dstRange[s%uint64(len(h.dstRange))]
}

// New creates a new prober
func New(c *config.Config, p config.Path, tos uint8) (*Prober, error) {
	pr := &Prober{
		cfg:           c,
		clock:         realClock{},
		hops:          confHopsToHops(c, p),
		path:          p,
		mtu:           mtuMax,
		transitProbes: newTransitProbes(),
		measurements:  measurement.NewDB(),
		srcAddrs:      generateAddrs(*c.SrcRange),
		tos:           tos,
		payload:       make(gopacket.Payload, *p.PayloadSizeBytes),
	}

	a, err := c.GetConfiguredSrcAddr()
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get configured source address")
	}

	pr.configuredSrcAddr = a
	return pr, nil
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

func (p *Prober) cleaner() {
	for {
		time.Sleep(time.Second)
		p.measurements.RemoveOlder(p.lastFinishedMeasurement())
	}
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
				srcRange: generateAddrs(*cfg.Routers[j].SrcRange),
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
