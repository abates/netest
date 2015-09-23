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
	sequence uint16
	socket   *net.UDPConn
}

func (c *Connection) getAddress(address string) *net.UDPAddr {
	var udpAddress *net.UDPAddr
	if c.err == nil {
		udpAddress, c.err = net.ResolveUDPAddr("udp", address)
		return udpAddress
	}
	return nil
}

func NewSink(addr string) (*Connection, error) {
	c := new(Connection)
	address := c.getAddress(addr)
	if c.err == nil {
		c.socket, c.err = net.ListenUDP("udp", address)
	}
	return c, c.err
}

func NewSrc(laddr, raddr string) (*Connection, error) {
	c := new(Connection)
	localAddr := c.getAddress(laddr)
	remoteAddr := c.getAddress(raddr)
	if c.err == nil {
		c.socket, c.err = net.DialUDP("udp", localAddr, remoteAddr)
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

	buf := make([]byte, 0)
	buffer := bytes.NewBuffer(buf)

	binary.Write(buffer, binary.BigEndian, h)
	buffer.Write(payload)
	_, c.err = buffer.WriteTo(c.socket)

	if c.err == nil {
		c.sequence++
	}
	return c.err
}

func (c *Connection) ReceiveMsg() (*Packet, error) {
	if c.err != nil {
		return nil, c.err
	}

	buf := make([]byte, MTU)
	_, c.err = c.socket.Read(buf)

	if c.err != nil {
		return nil, c.err
	}

	h := new(PacketHeader)
	buffer := bytes.NewBuffer(buf)
	binary.Read(buffer, binary.BigEndian, h)

	p := new(Packet)
	p.PacketHeader = *h
	if int(h.Length) > MTU {
		return p, fmt.Errorf("packet header indicates length of %d but it can be a maximum of %d", h.Length, MTU)
	}

	if h.Length-4 > 0 {
		p.Payload = buf[4 : h.Length-uint16(binary.Size(h))-1]
	} else {
		p.Payload = make([]byte, 0)
	}
	return p, c.err
}

func (c *Connection) Close() error {
	return c.socket.Close()
}
