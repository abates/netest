package netest

import (
	"io"
	"net"
	"sync"
)

type Sink struct {
	*Connection
	reader        io.Reader
	closeChannels []chan bool
	packets       chan *Packet
	wg            sync.WaitGroup
}

func NewUDPSink(laddr string) (*Sink, error) {
	c := new(Connection)
	s := new(Sink)
	s.packets = make(chan *Packet)
	s.Connection = c

	localAddr := c.getUDPAddr(laddr)

	var socket net.Conn
	if c.err == nil {
		socket, s.err = net.ListenUDP("udp", localAddr)
	}

	if s.err == nil {
		go s.receivePackets(socket, s.newCloseChannel())
	}
	return s, s.err
}

func NewTCPSink(laddr string) (*Sink, error) {
	c := new(Connection)
	s := new(Sink)
	s.packets = make(chan *Packet)
	s.Connection = c

	localAddr := c.getTCPAddr(laddr)

	var listener *net.TCPListener
	if c.err == nil {
		listener, s.err = net.ListenTCP("tcp", localAddr)
	}

	if s.err == nil {
		go func() {
			for {
				conn, err := listener.Accept()
				if err == nil {
					go s.receivePackets(conn, s.newCloseChannel())
				} else {
					logger.Warningf("Failed to accept new connection: %v", err)
				}
			}
		}()
	}
	return s, s.err
}

func (s *Sink) newCloseChannel() chan bool {
	closeChannel := make(chan bool)
	s.closeChannels = append(s.closeChannels, closeChannel)
	return closeChannel
}

func (s *Sink) ReceiveMsg() *Packet {
	return <-s.packets
}

func (s *Sink) receivePackets(conn net.Conn, closeChannel chan bool) {
	defer conn.Close()
	defer s.wg.Done()

	for {
		select {
		case <-closeChannel:
			break
		default:
			p, err := decodePacket(conn)
			if err == nil {
				s.packets <- p
			}
		}
	}
}

func (s *Sink) Close() {
	for _, closeCh := range s.closeChannels {
		closeCh <- true
	}
	s.wg.Wait()
}
