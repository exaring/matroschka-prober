package prober

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"sync/atomic"
	"time"

	"github.com/golang/glog"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"golang.org/x/net/ipv4"
)

// Prober represents a GRE/UDP packet prober
type Prober struct {
	srcRange              net.IPNet
	switchName            string
	deltaT                int64
	deltaTAggr            int64
	greDecapRange         net.IPNet
	mplsLabel             *uint32
	localAddress          map[string]net.IP
	configureLocalAddress net.IP
	rc                    *ipv4.RawConn
	udpc                  *net.UDPConn
	tos                   uint8
	tosName               string
	ttl                   uint8
	srcPort               uint16
	dstPort               uint16
	mtu                   uint16
	timeout               int64
	pps                   int
	transitProbes         *pMap
	measurements          *measurements
	measurementsAggr      *measurements
	probesSent            uint64
	probesReceived        uint64
	srcPOP                string
	dstPOP                string
	srcMetro              string
	dstMetro              string
	srcMachine            string
	spoofReplySrc         bool
	path                  []net.IPNet
	via                   []string
	payloadLen            int64
	payloadSer            gopacket.Payload
}

// Opts carries options for Prober
type Opts struct {
	SrcRange      net.IPNet
	SwitchName    string
	GREDecapRange net.IPNet
	MPLSLabel     *uint32
	LocalAddress  net.IP
	DeltaT        int64
	DeltaTAggr    int64
	Pps           int
	Tos           uint8
	TosName       string
	SrcPort       uint16
	DstPort       uint16
	Timeout       int64
	SrcPOP        string
	SrcMetro      string
	SrcMachine    string
	SpoofReplySrc bool
	Path          []net.IPNet // Used for explicit hops
	Via           []string    // Used for explicit hops labels
	PayloadLen    int64
}

type probe struct {
	Seq uint64
	Ts  int64
}

func (p *Prober) GetVia() []string {
	return p.via
}

func (p *Prober) GetSwitchName() string {
	return p.switchName
}

func (p *Prober) GetTOS() uint8 {
	return p.tos
}

func (p *Prober) GetGREDecapRange() net.IPNet {
	return p.greDecapRange
}

func (p *Prober) GetSrcRange() net.IPNet {
	return p.srcRange
}

func (p *Prober) GetDeltaT() int64 {
	return p.deltaT
}

func (p *Prober) GetDeltaTAggr() int64 {
	return p.deltaTAggr
}

func (p *Prober) GetLocalAddr(dst net.IP) net.IP {
	return p.localAddress[string(dst)]
}

func (p *Prober) GetTimeout() int64 {
	return p.timeout
}

func (p *Prober) GetPPS() int {
	return p.pps
}

func (p *Prober) GetTOSName() string {
	return p.tosName
}

func (p *Prober) GetSrcPOP() string {
	return p.srcPOP
}

func (p *Prober) GetDstPOP() string {
	return p.dstPOP
}

func (p *Prober) GetSrcMetro() string {
	return p.srcMetro
}

func (p *Prober) GetDstMetro() string {
	return p.dstMetro
}

func (p *Prober) GetSrcMachine() string {
	return p.srcMachine
}

func (p *Prober) GetMeasurement() *Measurement {
	now := time.Now().UnixNano()
	currentWindowBeginning := now - (now % p.deltaT)
	ts := int64(0)

	if now-currentWindowBeginning >= p.timeout {
		ts = currentWindowBeginning - p.deltaT
	} else {
		ts = currentWindowBeginning - 2*p.deltaT
	}

	m := p.measurements.get(ts)
	if m == nil {
		return nil
	}

	return m
}

func (p *Prober) GetAggregatedMeasurement() *Measurement {
	now := time.Now().UnixNano()
	currentWindowBeginning := now - (now % p.deltaTAggr)
	ts := int64(0)

	if now-currentWindowBeginning >= p.timeout {
		ts = currentWindowBeginning - p.deltaTAggr
	} else {
		ts = currentWindowBeginning - 2*p.deltaTAggr
	}

	m := p.measurementsAggr.get(ts)
	if m == nil {
		return nil
	}

	return m
}

func GetPOP(hostname string) (string, error) {
	parts := strings.Split(hostname, "-")
	if len(parts) < 2 {
		return "", fmt.Errorf("Invalid name: %s", hostname)
	}

	return parts[0], nil
}

func GetMetro(hostname string) (string, error) {
	parts := strings.Split(hostname, "-")
	if len(parts) < 2 {
		return "", fmt.Errorf("Invalid name: %s", hostname)
	}

	return parts[0][:3], nil
}

func getLocalAddr(dest net.IP) (net.IP, error) {
	conn, err := net.Dial("udp", fmt.Sprintf("%s:123", dest.String()))
	if err != nil {
		return nil, fmt.Errorf("Dial failed: %v", err)
	}

	conn.Close()

	host, _, err := net.SplitHostPort(conn.LocalAddr().String())
	if err != nil {
		return nil, fmt.Errorf("Unable to split host and port: %v", err)
	}

	return net.ParseIP(host), nil
}

// New creates and startes a new prober instance
func New(opts Opts) (*Prober, error) {
	if opts.Pps > 100000 {
		return nil, fmt.Errorf("Packet rate %d pps not supported (max 100000pps)", opts.Pps)
	}

	dstMetro, err := GetMetro(opts.SwitchName)
	if err != nil {
		return nil, fmt.Errorf("Unable to get metro name from switch name %s", opts.SwitchName)
	}

	dstPOP, err := GetPOP(opts.SwitchName)
	if err != nil {
		return nil, fmt.Errorf("Unable to get POP name from switch name %s", opts.SwitchName)
	}

	payload, err := createPayload(opts.PayloadLen)
	if err != nil {
		return nil, fmt.Errorf("Unable to create payload")
	}

	p := &Prober{
		switchName:            opts.SwitchName,
		srcRange:              opts.SrcRange,
		deltaT:                opts.DeltaT,
		deltaTAggr:            opts.DeltaTAggr,
		greDecapRange:         opts.GREDecapRange,
		mplsLabel:             opts.MPLSLabel,
		configureLocalAddress: opts.LocalAddress,
		tos:              opts.Tos,
		tosName:          opts.TosName,
		ttl:              64,
		srcPort:          opts.SrcPort,
		dstPort:          opts.DstPort,
		mtu:              1500,
		timeout:          opts.Timeout,
		pps:              opts.Pps,
		transitProbes:    newpMap(),
		measurements:     newMeasurements(),
		measurementsAggr: newMeasurements(),
		dstMetro:         dstMetro,
		dstPOP:           dstPOP,
		srcPOP:           opts.SrcPOP,
		srcMetro:         opts.SrcMetro,
		srcMachine:       opts.SrcMachine,
		spoofReplySrc:    opts.SpoofReplySrc,
		path:             opts.Path,
		via:              opts.Via,
		payloadLen:       opts.PayloadLen,
		payloadSer:       payload,
		localAddress:     make(map[string]net.IP),
	}

	err = p.initSockets()
	if err != nil {
		return nil, err
	}

	baseAddrDst := p.getDstRangeBase()
	maxIteratorDst := p.maskAddrCountDst()
	for i := uint32(0); i < maxIteratorDst; i++ {
		dst := net.IP(uint32Byte(baseAddrDst + uint32(i)%maxIteratorDst))

		if opts.LocalAddress != nil {
			p.localAddress[string(dst)] = opts.LocalAddress
		} else {
			src, err := getLocalAddr(dst)
			if err != nil {
				return nil, fmt.Errorf("Failed to determine return address for packets from %s to %s", src.String(), dst.String())
			}

			p.localAddress[string(dst)] = src
		}
	}

	go p.measurementsTimeoutChecker()
	go p.measurementsAggrTimeoutChecker()
	go p.rttTimeoutChecker()
	go p.sender()
	go p.receiver()

	return p, nil
}

func (p *Prober) initSockets() error {
	c, err := net.ListenPacket("ip4:47", "0.0.0.0") // GRE for IPv4
	if err != nil {
		return fmt.Errorf("Unable to listen for GRE packets: %v", err)
	}

	rc, err := ipv4.NewRawConn(c)
	if err != nil {
		return fmt.Errorf("Unable to create raw connection: %v", err)
	}

	p.rc = rc

	var udpConn *net.UDPConn

	// Try to find a free UDP port
	for {
		udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", p.dstPort))
		if err != nil {
			return fmt.Errorf("Unable to resolve address: %v", err)
		}

		udpConn, err = net.ListenUDP("udp", udpAddr)
		if err != nil {
			glog.Infof("UDP port %d is busy. Trying next one.", p.dstPort)
			p.dstPort++
			if p.dstPort > 65535 {
				return fmt.Errorf("Unable to listen for UDP packets: %v", err)
			}
			continue
		}
		break
	}

	p.udpc = udpConn
	return nil
}

func (p *Prober) measurementsTimeoutChecker() {
	for {
		wait := time.After(time.Duration(p.deltaT))
		<-wait

		now := time.Now().UnixNano()
		min := now - 3*p.deltaT

		p.measurements.removeOlder(min)
	}
}

func (p *Prober) measurementsAggrTimeoutChecker() {
	for {
		wait := time.After(time.Duration(p.deltaTAggr))
		<-wait

		now := time.Now().UnixNano()
		min := now - 3*p.deltaTAggr

		p.measurementsAggr.removeOlder(min)
	}
}

func (p *Prober) rttTimeoutChecker() {
	for {
		wait := time.After(time.Nanosecond * time.Duration(p.timeout/2))
		<-wait
		top := p.transitProbes.getLt(time.Now().UnixNano() - p.timeout)
		for s := range top {
			err := p.transitProbes.remove(s)
			if err != nil {
				glog.Infof("Probe %d timeouted: Unable to remove: %v", s, err)
			}
		}
	}
}

func (p *Prober) receiver() {
	for {
		recv := make([]byte, p.mtu)
		_, err := p.udpc.Read(recv)
		now := time.Now().UnixNano()
		if err != nil {
			glog.Errorf("Unable to read from UDP socket: %v", err)
		}

		pkt, err := unmarshal(recv)
		if err != nil {
			glog.Errorf("Unable to unmarshal message: %v", err)
		}

		err = p.transitProbes.remove(pkt.Seq)
		if err != nil {
			// Probe was count as lost, so we ignore it from here on
			continue
		}

		rtt := now - pkt.Ts
		if rtt > p.timeout {
			// Probe arrived late. rttTimoutChecker() will clean up after it. So we ignore it from here on
			continue
		}

		p.measurements.addRecv(pkt.Ts, rtt, p.deltaT)
		p.measurementsAggr.addRecv(pkt.Ts, rtt, p.deltaTAggr)
		atomic.AddUint64(&p.probesReceived, 1)
	}
}

func uint32b(data []byte) (ret uint32) {
	buf := bytes.NewBuffer(data)
	binary.Read(buf, binary.BigEndian, &ret)
	return
}

func uint32Byte(data uint32) (ret []byte) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, data)
	return buf.Bytes()
}

func (p *Prober) maskAddrCountSrc() uint32 {
	return maskAddrCount(p.srcRange)
}

func (p *Prober) maskAddrCountDst() uint32 {
	return maskAddrCount(p.greDecapRange)
}

func maskAddrCount(n net.IPNet) uint32 {
	ones, bits := n.Mask.Size()
	if ones == bits {
		return 1
	}

	x := uint32(1)
	for i := ones; i < bits; i++ {
		x = x * 2
	}
	return x
}

func getCIDRBase(n net.IPNet) uint32 {
	return uint32b(n.IP)
}

func (p *Prober) getSrcRangeBase() uint32 {
	return uint32b(p.srcRange.IP)
}

func (p *Prober) getDstRangeBase() uint32 {
	return uint32b(p.greDecapRange.IP)
}

func random(max int64) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(int(max))
}

func getNthAddr(n net.IPNet, i uint32) net.IP {
	baseAddr := getCIDRBase(n)
	c := maskAddrCount(n)
	return net.IP(uint32Byte(baseAddr + i%c))
}

func (p *Prober) sender() {
	seq := uint64(0)
	pr := probe{}

	// Desynchronize start times
	wait := time.After(time.Duration(random(p.timeout)) * time.Nanosecond)
	<-wait

	for {
		wait := time.After(time.Duration(1000000/p.pps) * time.Microsecond)
		<-wait

		src := getNthAddr(p.srcRange, uint32(seq))
		dst := getNthAddr(p.greDecapRange, uint32(seq))
		firstHop := dst

		hops := make([]net.IPNet, 0)
		if p.path != nil {
			firstHop = getNthAddr(p.path[0], uint32(seq))
			if len(p.path) > 1 {
				hops = append(hops, p.path[1:]...)
			}
			hops = append(hops, p.greDecapRange)
		}

		pr.Seq = seq
		pr.Ts = time.Now().UnixNano()
		pkt, err := p.craftPacket(&pr, src, dst, hops, firstHop, p.mplsLabel)
		if err != nil {
			glog.Errorf("Unable to craft packet: %v", err)
			continue
		}

		p.transitProbes.add(&pr)

		tsAligned := pr.Ts - (pr.Ts % p.deltaT)
		p.measurements.addSent(tsAligned)

		tsAlignedAggr := pr.Ts - (pr.Ts % p.deltaTAggr)
		p.measurementsAggr.addSent(tsAlignedAggr)

		err = p.sendPacket(pkt, src, firstHop)
		if err != nil {
			glog.Errorf("Unable to send packet: %v", err)
			p.transitProbes.remove(pr.Seq)
			continue
		}

		atomic.AddUint64(&p.probesSent, 1)
		seq++
	}
}

func unmarshal(data []byte) (*probe, error) {
	p := &probe{}
	err := binary.Read(bytes.NewReader(data), binary.BigEndian, p)
	if err != nil {
		return nil, fmt.Errorf("Unable to unmarshal read packet: %v", err)
	}

	return p, nil
}

func (p *probe) marshal() ([]byte, error) {
	var b bytes.Buffer
	err := binary.Write(&b, binary.BigEndian, p)
	if err != nil {
		return nil, fmt.Errorf("Unable to marshal: %v", err)
	}

	return b.Bytes(), nil
}

func createPayload(length int64) ([]byte, error) {
	payload := make([]byte, 0)
	for i := int64(0); i < length; i++ {
		payload = append(payload, 0x00)
	}

	var b bytes.Buffer
	err := binary.Write(&b, binary.BigEndian, payload)
	if err != nil {
		return nil, fmt.Errorf("Unable to marshal: %v", err)
	}

	return b.Bytes(), nil
}

func (p *Prober) craftPacket(pr *probe, src net.IP, dst net.IP, hops []net.IPNet, firstHop net.IP, label *uint32) ([]byte, error) {
	probeSer, err := pr.marshal()
	if err != nil {
		return nil, fmt.Errorf("Unable to marshal probe: %v", err)
	}

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}

	l := make([]gopacket.SerializableLayer, 0)

	greProto := layers.EthernetTypeIPv4
	if label != nil {
		greProto = layers.EthernetTypeMPLSUnicast
	}

	gre := &layers.GRE{
		Protocol: greProto,
	}
	l = append(l, gre)

	if label != nil {
		mpls := &layers.MPLS{
			Label:        *label,
			StackBottom:  true,
			TrafficClass: p.tos,
			TTL:          p.ttl,
		}
		l = append(l, mpls)
	}

	innerSrc := src
	if !p.spoofReplySrc {
		innerSrc = dst
	}

	lastHop := firstHop
	// Explicit hops
	for _, hop := range hops {
		dst := getNthAddr(hop, uint32(pr.Seq))
		l = append(l, &layers.IPv4{
			SrcIP:    src,
			DstIP:    dst,
			Version:  4,
			Protocol: layers.IPProtocolGRE,
			TOS:      p.tos,
			TTL:      p.ttl,
		})
		l = append(l, &layers.GRE{
			Protocol: layers.EthernetTypeIPv4,
		})
		lastHop = dst
	}

	// Create final UDP packet that will return
	ip := &layers.IPv4{
		SrcIP:    innerSrc,
		DstIP:    p.localAddress[string(lastHop)],
		Version:  4,
		Protocol: layers.IPProtocolUDP,
		TOS:      p.tos,
		TTL:      p.ttl,
	}
	l = append(l, ip)

	udp := &layers.UDP{
		SrcPort: layers.UDPPort(p.srcPort),
		DstPort: layers.UDPPort(p.dstPort),
	}

	udp.SetNetworkLayerForChecksum(ip)
	l = append(l, udp)
	l = append(l, gopacket.Payload(probeSer))
	l = append(l, p.payloadSer)

	// Serialize packet
	err = gopacket.SerializeLayers(buf, opts, l...)
	if err != nil {
		return nil, fmt.Errorf("Unable to serialize layers: %v", err)
	}

	packetData := buf.Bytes()

	return packetData, nil
}

func (p *Prober) sendPacket(payload []byte, src net.IP, firstHop net.IP) error {
	iph := &ipv4.Header{
		Src:      src,
		Dst:      firstHop,
		Version:  ipv4.Version,
		Len:      ipv4.HeaderLen,
		TOS:      int(p.tos),
		TotalLen: ipv4.HeaderLen + len(payload),
		TTL:      int(p.ttl),
		Protocol: 47,
	}

	// Set source IP on socket in order to enforce "ip rule..." rules (possible Linux bug)
	cm := ipv4.ControlMessage{}
	if p.configureLocalAddress != nil {
		cm.Src = p.configureLocalAddress
	}

	if err := p.rc.WriteTo(iph, payload, &cm); err != nil {
		return fmt.Errorf("Unable to send packet: %v", err)
	}

	return nil
}
