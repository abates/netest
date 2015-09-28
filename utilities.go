package netest

import (
	"fmt"
	"github.com/alexcesaro/log"
	"github.com/alexcesaro/log/golog"
	"net"
	"os"
)

const (
	Kilobyte = 1024.0
	Megabyte = 1048576.0
	Gigabyte = 1073741824.0
)

type Connection struct {
	err error
}

func (c *Connection) getUDPAddr(address string) *net.UDPAddr {
	var udpAddress *net.UDPAddr
	if c.err == nil {
		udpAddress, c.err = net.ResolveUDPAddr("udp", address)
		return udpAddress
	}
	return nil
}

func (c *Connection) getTCPAddr(address string) *net.TCPAddr {
	var tcpAddress *net.TCPAddr
	if c.err == nil {
		tcpAddress, c.err = net.ResolveTCPAddr("tcp", address)
		return tcpAddress
	}
	return nil
}
func Humanize(v float64) string {
	if v >= Gigabyte {
		return fmt.Sprintf("%.1f GB", (v / Gigabyte))
	} else if v >= Megabyte {
		return fmt.Sprintf("%.1f MB", (v / Megabyte))
	} else if v >= Kilobyte {
		return fmt.Sprintf("%.1f KB", (v / Kilobyte))
	} else {
		return fmt.Sprintf("%.1f B", v)
	}
}

var logger log.Logger

func init() {
	logger = golog.New(os.Stderr, log.Info)
}
