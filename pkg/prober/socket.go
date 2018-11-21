package prober

import (
	"fmt"
	"net"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/ipv4"
)

const (
	maxPort = uint16(65535)
)

type rawSocket interface {
	WriteTo(*ipv4.Header, []byte, *ipv4.ControlMessage) error
	Close() error
}

type udpSocket interface {
	Read([]byte) (int, error)
	Close() error
}

type rawSockWrapper struct {
	rawConn *ipv4.RawConn
}

func newRawSockWrapper() (*rawSockWrapper, error) {
	c, err := net.ListenPacket("ip4:47", "0.0.0.0") // GRE for IPv4
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

func (s *rawSockWrapper) WriteTo(h *ipv4.Header, p []byte, cm *ipv4.ControlMessage) error {
	return s.rawConn.WriteTo(h, p, cm)
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
			log.Debugf("UDP port %d is busy. Trying next one.", port)
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
	rc, err := newRawSockWrapper()
	if err != nil {
		return fmt.Errorf("Unable to create rack socket wrapper: %v", err)
	}

	p.rawConn = rc
	return nil
}

func (p *Prober) initUDPSocket() error {
	s, err := newUDPSockWrapper(*p.cfg.BasePort)
	if err != nil {
		return fmt.Errorf("Unable to get UDP socket wrapper: %v", err)
	}

	p.udpConn = s
	p.dstUDPPort = s.getPort()
	return nil
}

func (p *Prober) setLocalAddr() error {
	addr, err := getLocalAddr(p.hops[0].dstRange[0])
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
