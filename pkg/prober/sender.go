package prober

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/ipv4"
)

func (p *Prober) sender() {
	defer p.rawConn.Close()

	p.desynchronizeStartTime()
	p.setLocalAddr()
	seq := uint64(0)
	pr := probe{}
	t := time.NewTicker(time.Second / time.Duration(*p.path.PPS))

	for {
		select {
		case <-p.stop:
			return
		case <-t.C:
		}

		pr.Seq = seq
		pr.Ts = time.Now().UnixNano()
		pkt, err := p.craftPacket(&pr)
		if err != nil {
			log.Errorf("Unable to craft packet: %v", err)
			continue
		}

		p.transitProbes.add(&pr)

		tsAligned := pr.Ts - (pr.Ts % (int64(*p.path.MeasurementLengthMS) * int64(time.Millisecond)))
		p.measurements.AddSent(tsAligned)

		srcAddr := p.getSrcAddr(seq)
		dstAddr := p.hops[0].getAddr(seq)
		err = p.sendPacket(pkt, srcAddr, dstAddr)
		if err != nil {
			log.Errorf("Unable to send packet: %v", err)
			p.transitProbes.remove(pr.Seq)
			continue
		}

		atomic.AddUint64(&p.probesSent, 1)
		seq++
	}
}

func (p *Prober) sendPacket(payload []byte, src net.IP, dst net.IP) error {
	iph := &ipv4.Header{
		Src:      src,
		Dst:      dst,
		Version:  ipv4.Version,
		Len:      ipv4.HeaderLen,
		TOS:      int(p.tos),
		TotalLen: ipv4.HeaderLen + len(payload),
		TTL:      ttl,
		Protocol: 47, //GRE
	}

	// Set source IP on socket in order to enforce "ip rule..." rules (possible Linux bug)
	cm := ipv4.ControlMessage{}

	if err := p.rawConn.WriteTo(iph, payload, &cm); err != nil {
		return fmt.Errorf("Unable to send packet: %v", err)
	}

	return nil
}

func (p *Prober) desynchronizeStartTime() {
	time.Sleep(time.Duration(random(int64(*p.path.TimeoutMS))) * time.Microsecond)
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

func uint32Byte(data uint32) (ret []byte) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, data)
	return buf.Bytes()
}
