package netest

import (
	"bytes"
	"encoding/binary"
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
	h := new(PacketHeader)
	h.Length = uint16(len(payload) + 4)
	h.Sequence = c.sequence

	buf := make([]byte, h.Length)
	buffer := bytes.NewBuffer(buf)

	binary.Write(buffer, binary.BigEndian, h)
	buffer.Write(payload)
	_, c.err = buffer.WriteTo(c.socket)

	if c.err != nil {
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
	if h.Length-4 > 0 {
		p.Payload = buf[4 : h.Length-1]
	} else {
		p.Payload = make([]byte, 0)
	}
	return p, c.err
}

func (c *Connection) Close() error {
	return c.socket.Close()
}
