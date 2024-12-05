package prober

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type probe struct {
	SequenceNumber uint64
	TimeStamp     int64
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
