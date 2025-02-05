package prober

import (
	"fmt"
	"net"
	"strconv"

	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
	"golang.org/x/sys/unix"
)

const (
	maxPort = uint16(65535)
)

type rawSocket interface {
	WriteTo(payload []byte, options writeOptions) error
	Close() error
}

type writeOptions struct {
	src      net.IP
	dst      net.IP
	tos      int64
	ttl      int64
	protocol int64
}

type udpSocket interface {
	Read([]byte) (int, error)
	Close() error
}

type rawSockWrapper struct {
	rawConn *ipv4.RawConn
}

func newRawSockWrapper() (*rawSockWrapper, error) {
	greProtoStr := strconv.FormatInt(unix.IPPROTO_GRE, 10)
	c, err := net.ListenPacket("ip4:"+greProtoStr, "0.0.0.0") // GRE for IPv4
	if err != nil {
		return nil, fmt.Errorf("Unable to listen for GRE packets: %v", err)
	}

	rc, err := ipv4.NewRawConn(c)
	if err != nil {
		return nil, fmt.Errorf("Unable to create raw connection: %v", err)
	}

	return &rawSockWrapper{
		rawConn: rc,
	}, nil
}

func (s *rawSockWrapper) WriteTo(p []byte, o writeOptions) error {

	iph := &ipv4.Header{
		Src:      o.src,
		Dst:      o.dst,
		Version:  ipv4.Version,
		Len:      ipv4.HeaderLen,
		TOS:      int(o.tos),
		TotalLen: ipv4.HeaderLen + len(p),
		TTL:      ttl,
		Protocol: unix.IPPROTO_GRE,
	}
	cm := &ipv4.ControlMessage{}
	if o.src != nil {
		cm.Src = o.src
	}

	return s.rawConn.WriteTo(iph, p, cm)
}

func (s *rawSockWrapper) Close() error {
	return s.rawConn.Close()
}

type udpSockWrapper struct {
	udpConn *net.UDPConn
	port    uint16
}

func newUDPSockWrapper(basePort uint16) (*udpSockWrapper, error) {
	var udpConn *net.UDPConn

	port := basePort
	// Try to find a free UDP port
	for {
		udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
		if err != nil {
			return nil, fmt.Errorf("Unable to resolve address: %v", err)
		}

		udpConn, err = net.ListenUDP("udp", udpAddr)
		if err != nil {
			//log.Debugf("UDP port %d is busy. Trying next one.", port)
			port++
			if port > maxPort {
				return nil, fmt.Errorf("Unable to listen for UDP packets: %v", err)
			}
			continue
		}
		break
	}

	return &udpSockWrapper{
		udpConn: udpConn,
		port:    port,
	}, nil
}

func (u *udpSockWrapper) getPort() uint16 {
	return u.port
}

func (u *udpSockWrapper) Read(b []byte) (int, error) {
	return u.udpConn.Read(b)
}

func (u *udpSockWrapper) Close() error {
	return u.udpConn.Close()
}

func (p *Prober) initRawSocket() error {
	ipVersion := p.cfg.IPProtocol

	if ipVersion == 4 {
		rc, err := newRawSockWrapper()
		if err != nil {
			return fmt.Errorf("Unable to create rack socket wrapper: %v", err)
		}

		p.rawConn = rc
	}

	if ipVersion == 6 {
		rc, err := newIPv6RawSockWrapper()
		if err != nil {
			return fmt.Errorf("Unable to create rack socket wrapper: %v", err)
		}

		p.rawConn = rc
	}

	return nil
}

func (p *Prober) initUDPSocket() error {
	s, err := newUDPSockWrapper(p.cfg.BasePort)
	if err != nil {
		return fmt.Errorf("Unable to get UDP socket wrapper: %v", err)
	}

	p.udpConn = s
	p.dstUDPPort = s.getPort()
	return nil
}

func (p *Prober) setLocalAddr() error {
	if p.cfg.ConfiguredSrcAddr != nil {
		p.localAddr = p.cfg.ConfiguredSrcAddr
		return nil
	}

	addr, err := getLocalAddr(p.cfg.Hops[0].DstRange[0])
	if err != nil {
		return fmt.Errorf("Unable to get local address: %v", err)
	}

	p.localAddr = addr
	return nil
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

type rawIPv6SocketWrapper struct {
	rawIPv6Conn *ipv6.PacketConn
}

func (s *rawIPv6SocketWrapper) WriteTo(p []byte, o writeOptions) error {
	cm := &ipv6.ControlMessage{
		TrafficClass: int(o.tos),
		HopLimit:     ttl,
		Src:          o.src,
		Dst:          o.dst,
	}

	dstAddress := net.IPAddr{IP: o.dst}

	_, err := s.rawIPv6Conn.WriteTo(p, cm, &dstAddress)
	return err
}

func (s *rawIPv6SocketWrapper) Close() error {
	return s.rawIPv6Conn.Close()
}

func newIPv6RawSockWrapper() (*rawIPv6SocketWrapper, error) {
	greProtoStr := strconv.FormatInt(unix.IPPROTO_GRE, 10)
	c, err := net.ListenPacket("ip6:"+greProtoStr, "::")
	if err != nil {
		return nil, fmt.Errorf("Unable to listen for GRE packets: %v", err)
	}

	rc := ipv6.NewPacketConn(c)

	return &rawIPv6SocketWrapper{
		rawIPv6Conn: rc,
	}, nil
}
