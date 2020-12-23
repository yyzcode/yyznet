package protocol

import (
	"net"
)

type Tcp struct {
}

func (t Tcp) ReadOnePkg(conn *net.Conn) (pkg []byte, err error) {
	var buffer [1024 * 1024]byte
	n, err := (*conn).Read(buffer[:])
	pkg = buffer[:n]
	return
}

func (t Tcp) EncodeOnePkg(data string) []byte {
	return []byte(data)
}
