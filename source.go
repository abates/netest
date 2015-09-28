package netest

import (
	"bufio"
	"net"
)

type Source struct {
	*Connection
	writer   *bufio.Writer
	sequence uint16
	socket   net.Conn
}

func NewUDPSource(laddr, raddr string) (*Source, error) {
	c := new(Connection)
	s := new(Source)
	s.Connection = c

	localAddr := c.getUDPAddr(laddr)
	remoteAddr := c.getUDPAddr(raddr)

	if c.err == nil {
		s.socket, s.err = net.DialUDP("udp", localAddr, remoteAddr)
	}

	if s.err == nil {
		s.writer = bufio.NewWriter(s.socket)
	}
	return s, s.err
}

func NewTCPSource(laddr, raddr string) (*Source, error) {
	c := new(Connection)
	s := new(Source)
	s.Connection = c

	localAddr := c.getTCPAddr(laddr)
	remoteAddr := c.getTCPAddr(raddr)

	if c.err == nil {
		s.socket, s.err = net.DialTCP("tcp", localAddr, remoteAddr)
	}

	if s.err == nil {
		s.writer = bufio.NewWriter(s.socket)
	}

	return s, s.err
}

func (s *Source) SendMsg(payload []byte) (int, error) {
	var length int
	length, s.err = encodePacket(s.writer, s.sequence, payload)

	if s.err == nil {
		s.sequence++
	}
	return length, s.err
}

func (s *Source) Close() {
	s.socket.Close()
}
