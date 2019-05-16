package prober

import (
	"fmt"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

const (
	ttl = 64
)

func (p *Prober) getSrcAddrHop(hop int, seq uint64) net.IP {
	return p.hops[hop-1].srcRange[seq%uint64(len(p.hops[hop-1].srcRange))]
}

func (p *Prober) getDstAddr(hop int, seq uint64) net.IP {
	return p.hops[hop].dstRange[seq%uint64(len(p.hops[hop].dstRange))]
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

	l := make([]gopacket.SerializableLayer, 0, (len(p.hops)-1)*2+5)
	l = append(l, &layers.GRE{
		Protocol: layers.EthernetTypeIPv4,
	})

	for i := range p.hops {
		if i == 0 {
			continue
		}

		l = append(l, &layers.IPv4{
			SrcIP:    p.getSrcAddrHop(i, pr.Seq),
			DstIP:    p.getDstAddr(i, pr.Seq),
			Version:  4,
			Protocol: layers.IPProtocolGRE,
			TOS:      p.tos,
			TTL:      ttl,
		})

		l = append(l, &layers.GRE{
			Protocol: layers.EthernetTypeIPv4,
		})
	}

	// Create final UDP packet that will return
	ip := &layers.IPv4{
		SrcIP:    p.getSrcAddrHop(len(p.hops), pr.Seq),
		DstIP:    p.localAddr,
		Version:  4,
		Protocol: layers.IPProtocolUDP,
		TOS:      p.tos,
		TTL:      ttl,
	}
	l = append(l, ip)

	udp := &layers.UDP{
		SrcPort: layers.UDPPort(p.dstUDPPort),
		DstPort: layers.UDPPort(p.dstUDPPort),
	}

	udp.SetNetworkLayerForChecksum(ip)
	l = append(l, udp)
	l = append(l, gopacket.Payload(probeSer))
	l = append(l, p.payload)

	err = gopacket.SerializeLayers(buf, opts, l...)
	if err != nil {
		return nil, fmt.Errorf("Unable to serialize layers: %v", err)
	}

	return buf.Bytes(), nil

}
