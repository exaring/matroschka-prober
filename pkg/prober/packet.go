package prober

import (
	"fmt"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

const (
	ttl = 64
)

func (p *Prober) craftPacket(pr *probe) ([]byte, error) {
	srcAddr := p.srcAddrs[pr.Seq%uint64(len(p.srcAddrs))]

	probeSer, err := pr.marshal()
	if err != nil {
		return nil, fmt.Errorf("Unable to marshal probe: %v", err)
	}

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}

	innerSrc := srcAddr
	if !*p.path.SpoofReplySrc {
		lastHop := p.hops[len(p.hops)-1]
		innerSrc = lastHop.dstRange[len(lastHop.dstRange)]
	}

	greHeader := &layers.GRE{
		Protocol: layers.EthernetTypeIPv4,
	}
	if p.cfg.Defaults.KeyId != nil {
		greHeader.Key = *p.cfg.Defaults.KeyId
		greHeader.KeyPresent = true
	}

	l := make([]gopacket.SerializableLayer, 0, (len(p.hops)-1)*2+5)
	l = append(l, greHeader)

	for i := range p.hops {
		if i == 0 {
			continue
		}

		dstAddr := p.hops[i].dstRange[pr.Seq%uint64(len(p.hops[i].dstRange))]
		l = append(l, &layers.IPv4{
			SrcIP:    srcAddr,
			DstIP:    dstAddr,
			Version:  4,
			Protocol: layers.IPProtocolGRE,
			TOS:      p.tos,
			TTL:      ttl,
		})

		l = append(l, greHeader)
	}

	// Create final UDP packet that will return
	ip := &layers.IPv4{
		SrcIP:    innerSrc,
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
