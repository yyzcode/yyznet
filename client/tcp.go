package client

import (
	"fmt"
	"github.com/yyzcoder/yyznet/protocol"
	"net"
	"sync"
)

type Tcp struct {
	Addr      string
	Port      int
	conn      net.Conn
	Status    int //状态0连接中，状态1连接成功，状态2已断开
	Protocol  protocol.Tcp
	OnConnect func()
	OnMessage func([]byte)
	OnClose   func()
}

var TcpWg sync.WaitGroup

func (t *Tcp) Exec() {
	var err error
	t.conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", t.Addr, t.Port))
	TcpWg.Done()
	if err != nil {
		t.Status = 2
		fmt.Println("本地连接错误：", err)
		return
	}
	t.Status = 1
	if t.OnConnect != nil {
		t.OnConnect()
	}
	for {
		data, err := t.Protocol.ReadOnePkg(&t.conn)
		if len(data) > 0 {
			t.OnMessage(data)
		}
		if err != nil {
			if t.OnClose != nil {
				t.OnClose()
			}
			break
		}
	}
}
func (t *Tcp) Send(d []byte) {
	_, err := t.conn.Write(d)
	if err != nil {
		fmt.Println("try to send data to TcpClient fail:", err)
	}
	return
}
func (t *Tcp) Close() {
	err := t.conn.Close()
	if err != nil {
		fmt.Println("try to close TcpClient fail:", err)
	}
	return
}
