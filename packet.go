package netest

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

var MTU = 1500

type PacketHeader struct {
	Length   uint16
	Sequence uint16
}

type Packet struct {
	PacketHeader
	Payload []byte
}

func decodePacket(reader io.Reader) (*Packet, error) {
	b := make([]byte, MTU)
	p := new(Packet)

	length, err := reader.Read(b)
	if err == nil {
		buffer := bytes.NewBuffer(b[0:length])
		h := new(PacketHeader)
		binary.Read(buffer, binary.BigEndian, h)
		p.PacketHeader = *h

		if int(h.Length) > length {
			return p, fmt.Errorf("Packet Header reports length %d which is larger than the buffer size of %d", h.Length, len(b))
		}

		if h.Length-4 > 0 {
			p.Payload = b[4:h.Length]
		} else {
			p.Payload = make([]byte, 0)
		}
	}
	return p, err
}

func encodePacket(writer *bufio.Writer, sequence uint16, payload []byte) (int, error) {
	var length int
	if len(payload) > MTU {
		return 0, fmt.Errorf("payload cannot be greater than %d bytes", MTU)
	}

	h := new(PacketHeader)
	h.Length = uint16(len(payload) + binary.Size(h))
	h.Sequence = sequence

	err := binary.Write(writer, binary.BigEndian, h)
	if err == nil {
		_, err = writer.Write(payload)
	}

	if err == nil {
		length = writer.Buffered()
		err = writer.Flush()
	}
	return length, err
}
