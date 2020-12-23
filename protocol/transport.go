package protocol

import "net"

type Transport interface {
	ReadOnePkg(conn *net.Conn) (pkg []byte, err error)
	EncodeOnePkg(data string) []byte
}
