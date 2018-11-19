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

func (p *Prober) initRawSocket() error {
	c, err := net.ListenPacket("ip4:47", "0.0.0.0") // GRE for IPv4
	if err != nil {
		return fmt.Errorf("Unable to listen for GRE packets: %v", err)
	}

	rc, err := ipv4.NewRawConn(c)
	if err != nil {
		return fmt.Errorf("Unable to create raw connection: %v", err)
	}

	p.rawConn = rc
	return nil
}

func (p *Prober) initUDPSocket() error {
	var udpConn *net.UDPConn

	p.dstUDPPort = *p.cfg.BasePort
	// Try to find a free UDP port
	for {
		udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", p.dstUDPPort))
		if err != nil {
			return fmt.Errorf("Unable to resolve address: %v", err)
		}

		udpConn, err = net.ListenUDP("udp", udpAddr)
		if err != nil {
			log.Debugf("UDP port %d is busy. Trying next one.", p.dstUDPPort)
			p.dstUDPPort++
			if p.dstUDPPort > maxPort {
				return fmt.Errorf("Unable to listen for UDP packets: %v", err)
			}
			continue
		}
		break
	}

	p.udpConn = udpConn
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
