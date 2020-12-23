package server

import (
	"fmt"
	"github.com/yyzcoder/yyznet/protocol"
	"net"
)

type Tcp struct {
	ListenAddr string
	ListenPort int
	Protocol   protocol.Transport
	OnStart    func()
	OnConnect  func(*Connect)
	OnMessage  func(*Connect, []byte)
	OnClose    func(*Connect)
}

type Connect struct {
	Id int
	net.Conn
}

var connId = 1

func (s *Tcp) Run() (err error) {
	if s.ListenPort > 65536 || s.ListenPort < 1 {
		err = fmt.Errorf("listen port mast between 1 and 65535")
	}
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.ListenAddr, s.ListenPort))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer listener.Close()
	if s.OnStart != nil {
		s.OnStart()
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		c := Connect{connId, conn}
		connId++
		if s.OnConnect != nil {
			s.OnConnect(&c)
		}
		go s.listenConn(&c)
	}
}

func (s *Tcp) listenConn(conn *Connect) {
	defer (*conn).Close()
	for {
		data, err := s.Protocol.ReadOnePkg(&conn.Conn)
		if len(data) > 0 && s.OnMessage != nil {
			s.OnMessage(conn, data)
		}
		if err != nil {
			if s.OnClose != nil {
				s.OnClose(conn)
			}
			conn.Close()
			break
		}
	}
}
