package netest

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
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

type Connection struct {
	err      error
	socket   net.Conn
	sequence uint16
	buffer   []byte
}

func (c *Connection) getUDPAddr(address string) *net.UDPAddr {
	var udpAddress *net.UDPAddr
	if c.err == nil {
		udpAddress, c.err = net.ResolveUDPAddr("udp", address)
		return udpAddress
	}
	return nil
}

func newConnection() *Connection {
	c := new(Connection)
	c.buffer = make([]byte, MTU)
	return c
}

func NewUdpSrc(laddr, raddr string) (*Connection, error) {
	c := newConnection()

	localAddr := c.getUDPAddr(laddr)
	remoteAddr := c.getUDPAddr(raddr)

	if c.err == nil {
		c.socket, c.err = net.DialUDP("udp", localAddr, remoteAddr)
	}

	return c, c.err
}

func NewUdpSink(laddr string) (*Connection, error) {
	c := newConnection()
	localAddr := c.getUDPAddr(laddr)

	if c.err == nil {
		c.socket, c.err = net.ListenUDP("udp", localAddr)
	}
	return c, c.err
}

func (c *Connection) SendMsg(payload []byte) error {
	if c.err != nil {
		return c.err
	}

	if len(payload) > MTU {
		return fmt.Errorf("payload cannot be greater than %d bytes", MTU)
	}

	h := new(PacketHeader)
	h.Length = uint16(len(payload) + binary.Size(h))
	h.Sequence = c.sequence
	buffer := bytes.NewBuffer(make([]byte, 0))

	c.err = binary.Write(buffer, binary.BigEndian, h)
	if c.err == nil {
		_, c.err = buffer.Write(payload)
		if c.err == nil {
			_, c.err = buffer.WriteTo(c.socket)
		}
	}

	if c.err == nil {
		c.sequence++
	}
	return c.err
}

func (c *Connection) ReceiveMsg() (*Packet, error) {
	if c.err != nil {
		return nil, c.err
	}
	p := new(Packet)

	//b := make([]byte, MTU)
	length, err := c.socket.Read(c.buffer)
	if err == nil {
		buffer := bytes.NewBuffer(c.buffer[0:length])
		h := new(PacketHeader)
		binary.Read(buffer, binary.BigEndian, h)
		p.PacketHeader = *h

		if int(h.Length) > length {
			return p, fmt.Errorf("Packet Header reports length %d which is larger than the buffer size of %d", h.Length, len(c.buffer))
		}

		if h.Length-4 > 0 {
			p.Payload = c.buffer[4:h.Length]
		} else {
			p.Payload = make([]byte, 0)
		}
	}
	return p, c.err
}

func (c *Connection) Close() error {
	return c.socket.Close()
}
