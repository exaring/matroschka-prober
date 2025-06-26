package prober

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"unsafe"
)

var (
	endianNess binary.ByteOrder
)

func init() {
	endianNess = getEndianNess()
}

func getEndianNess() binary.ByteOrder {
	a := binary.NativeEndian.Uint16([]byte{
		1, 0,
	})
	b := binary.LittleEndian.Uint16([]byte{
		1, 0,
	})

	if a == b {
		return binary.LittleEndian
	}

	return binary.BigEndian
}

type probe struct {
	SequenceNumber uint64
	TimeStamp      int64
}

func unmarshal(data []byte) (*probe, error) {
	p := &probe{}
	err := binary.Read(bytes.NewReader(data), binary.BigEndian, p)
	if err != nil {
		return nil, fmt.Errorf("Unable to unmarshal read packet: %v", err)
	}

	return p, nil
}

func (p *probe) marshal() [16]byte {
	sn := Uint64Byte(p.SequenceNumber)
	toBigEndian(sn[:])

	ts := Int64Byte(p.TimeStamp)
	toBigEndian(ts[:])

	return [16]byte{
		sn[0], sn[1], sn[2], sn[3], sn[4], sn[5], sn[6], sn[7],
		ts[0], ts[1], ts[2], ts[3], ts[4], ts[5], ts[6], ts[7],
	}
}

func Uint64Byte(x uint64) [8]byte {
	return *(*[8]byte)(unsafe.Pointer(&x))
}

func Int64Byte(x int64) [8]byte {
	return *(*[8]byte)(unsafe.Pointer(&x))
}

func toBigEndian(a []byte) {
	if endianNess == binary.BigEndian {
		return
	}

	for i := 0; i < len(a)/2; i++ {
		tmp := a[i]
		a[i] = a[len(a)-i-1]
		a[len(a)-i-1] = tmp
	}
}
