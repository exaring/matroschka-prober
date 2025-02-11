package prober

import (
	"fmt"
	"net"
	"strconv"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

const (
	ttl = 64
)

func (p *Prober) getSrcAddrHop(hop int, seq uint64) net.IP {
	return p.cfg.Hops[hop-1].SrcRange[seq%uint64(len(p.cfg.Hops[hop-1].SrcRange))]
}

func (p *Prober) getDstAddr(hop int, seq uint64) net.IP {
	return p.cfg.Hops[hop].DstRange[seq%uint64(len(p.cfg.Hops[hop].DstRange))]
}

func (p *Prober) getIPVersion() (int8, error) {
	firstHop := net.IP{}
	if len(p.cfg.Hops[0].SrcRange) > 0 {
		firstHop = p.cfg.Hops[0].SrcRange[0]
	}

	version := firstHop.To4()
	if version != nil {
		return 4, nil
	}

	version = firstHop.To16()
	if version != nil {
		return 6, nil
	}

	malformedAddress := ""
	for i := range firstHop {
		malformedAddress = malformedAddress + strconv.Itoa(int(firstHop[i])) + "."
	}
	return 0, fmt.Errorf("Couldn't determine the protocol version for address %s", malformedAddress)

}

func (p *Prober) craftIPV4Packet(pr *probe, l []gopacket.SerializableLayer) ([]gopacket.SerializableLayer, error) {
	l = append(l, &layers.GRE{
		Protocol: layers.EthernetTypeIPv4,
	})

	for i := range p.cfg.Hops {
		if i == 0 {
			continue
		}

		l = append(l, &layers.IPv4{
			SrcIP:    p.getSrcAddrHop(i, pr.SequenceNumber),
			DstIP:    p.getDstAddr(i, pr.SequenceNumber),
			Version:  4,
			Protocol: layers.IPProtocolGRE,
			TOS:      p.cfg.TOS.Value,
			TTL:      ttl,
		})

		l = append(l, &layers.GRE{
			Protocol: layers.EthernetTypeIPv4,
		})
	}

	// Create final UDP packet that will return
	ip := &layers.IPv4{
		SrcIP:    p.getSrcAddrHop(len(p.cfg.Hops), pr.SequenceNumber),
		DstIP:    p.localAddr,
		Version:  4,
		Protocol: layers.IPProtocolUDP,
		TOS:      p.cfg.TOS.Value,
		TTL:      ttl,
	}
	l = append(l, ip)

	udp := &layers.UDP{
		SrcPort: layers.UDPPort(p.dstUDPPort),
		DstPort: layers.UDPPort(p.dstUDPPort),
	}

	err := udp.SetNetworkLayerForChecksum(ip)
	if err != nil {
		return nil, fmt.Errorf("couldn't set the network layer for checksum: %w", err)
	}
	l = append(l, udp)

	return l, nil
}

func (p *Prober) craftIPV6Packet(pr *probe, l []gopacket.SerializableLayer) ([]gopacket.SerializableLayer, error) {
	l = append(l, &layers.GRE{
		Protocol: layers.EthernetTypeIPv6,
	})

	for i := range p.cfg.Hops {
		if i == 0 {
			continue
		}

		l = append(l, &layers.IPv6{
			SrcIP:        p.getSrcAddrHop(i, pr.SequenceNumber),
			DstIP:        p.getDstAddr(i, pr.SequenceNumber),
			Version:      6,
			TrafficClass: p.cfg.TOS.Value,
			NextHeader:   layers.IPProtocolGRE,
			HopLimit:     ttl,
		})

		l = append(l, &layers.GRE{
			Protocol: layers.EthernetTypeIPv6,
		})

	}

	ip := &layers.IPv6{
		SrcIP:        p.getSrcAddrHop(len(p.cfg.Hops), pr.SequenceNumber),
		DstIP:        p.localAddr,
		Version:      6,
		TrafficClass: p.cfg.TOS.Value,
		NextHeader:   layers.IPProtocolUDP,
		HopLimit:     ttl,
	}
	l = append(l, ip)

	udp := &layers.UDP{
		SrcPort: layers.UDPPort(p.dstUDPPort),
		DstPort: layers.UDPPort(p.dstUDPPort),
	}

	err := udp.SetNetworkLayerForChecksum(ip)
	if err != nil {
		return nil, fmt.Errorf("couldn't set the network layer for checksum: %w", err)
	}
	l = append(l, udp)

	return l, nil
}

func (p *Prober) craftPacket(pr *probe) ([]byte, error) {
	probeSer, err := pr.marshal()
	if err != nil {
		return nil, fmt.Errorf("Unable to marshal probe: %v", err)
	}

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}

	l := make([]gopacket.SerializableLayer, 0, (len(p.cfg.Hops)-1)*2+5)

	ipProtocolVersion := p.cfg.IPVersion
	if err != nil {
		return nil, err
	}

	if ipProtocolVersion == 4 {
		l, err = p.craftIPV4Packet(pr, l)
		if err != nil {
			return nil, fmt.Errorf("failed to craft IPv4 packet: %w", err)
		}
	}

	if ipProtocolVersion == 6 {
		l, err = p.craftIPV6Packet(pr, l)
		if err != nil {
			return nil, fmt.Errorf("failed to craft IPv6 packet: %w", err)
		}
	}

	l = append(l, gopacket.Payload(probeSer))
	l = append(l, p.payload)

	err = gopacket.SerializeLayers(buf, opts, l...)
	if err != nil {
		return nil, fmt.Errorf("Unable to serialize layers: %v", err)
	}

	return buf.Bytes(), nil
}
