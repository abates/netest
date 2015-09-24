package netest

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
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

type readWriter struct {
	reader *io.PipeReader
	writer *io.PipeWriter
	buffer *bytes.Buffer
}

func (rw *readWriter) Read(p []byte) (int, error) {
	return rw.reader.Read(p)
}

func (rw *readWriter) Write(p []byte) (int, error) {
	return rw.buffer.Write(p)
}

func (rw *readWriter) Flush() (int64, error) {
	return rw.buffer.WriteTo(rw.writer)
}

func (rw *readWriter) Close() error {
	rw.reader.Close()
	return rw.writer.Close()
}

type Connection struct {
	err      error
	socket   *readWriter
	sequence uint16
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
	c.socket = new(readWriter)
	c.socket.reader, c.socket.writer = io.Pipe()
	c.socket.buffer = bytes.NewBuffer(make([]byte, 0))
	return c
}

func NewUdpSrc(laddr, raddr string) (*Connection, error) {
	c := newConnection()

	localAddr := c.getUDPAddr(laddr)
	remoteAddr := c.getUDPAddr(raddr)

	var socket *net.UDPConn
	if c.err == nil {
		socket, c.err = net.DialUDP("udp", localAddr, remoteAddr)
	}

	if c.err == nil {
		go func() {
			_, err := io.Copy(socket, c.socket)
			if err != nil {
				panic(fmt.Sprintf("Failed to copy to the network: %v", err))
			}
			socket.Close()
		}()
	}
	return c, c.err
}

func NewUdpSink(laddr string) (*Connection, error) {
	c := new(Connection)
	localAddr := c.getUDPAddr(laddr)

	if c.err == nil {
		var socket *net.UDPConn
		socket, c.err = net.ListenUDP("udp", localAddr)
		if c.err == nil {
			go func() {
				buffer := make([]byte, MTU)
				io.CopyBuffer(socket, c.socket, buffer)
				socket.Close()
			}()
		}
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

	binary.Write(c.socket, binary.BigEndian, h)
	_, c.err = c.socket.Write(payload)
	if c.err == nil {
		_, c.err = c.socket.Flush()
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

	h := new(PacketHeader)
	binary.Read(c.socket, binary.BigEndian, h)

	p := new(Packet)
	p.PacketHeader = *h

	if h.Length-4 > 0 {
		p.Payload = make([]byte, h.Length)
		_, c.err = c.socket.Read(p.Payload)
	} else {
		p.Payload = make([]byte, 0)
	}
	return p, c.err
}

func (c *Connection) Close() error {
	return c.socket.Close()
}
