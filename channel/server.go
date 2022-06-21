package channel

import (
	"encoding/json"
	"fmt"
	"github.com/yyzcode/yyznet/protocol"
	"github.com/yyzcode/yyznet/server"
	"sync"
)

type Server struct {
	ListenAddr      string
	ListenPort      int
	protocol        protocol.Transport
	eventsSubscribe map[string][]*server.Connect
}

var lock sync.Mutex

func (s *Server) Run() (err error) {
	s.protocol = protocol.Frame{}
	s.eventsSubscribe = make(map[string][]*server.Connect, 0)
	tcpServer := server.Tcp{
		ListenPort: s.ListenPort,
		ListenAddr: s.ListenAddr,
		Protocol:   s.protocol,
	}
	tcpServer.OnStart = func() {
		fmt.Printf("channel server start successful %s:%d\n", s.ListenAddr, s.ListenPort)
	}
	tcpServer.OnConnect = func(conn *server.Connect) {

	}
	tcpServer.OnMessage = func(conn *server.Connect, data []byte) {
		bus := new(BusData)
		if err := json.Unmarshal(data, bus); err != nil {
			fmt.Println(err)
			return
		}
		switch bus.InfoType {
		case "subscribe":
			s.subscribe(bus.Event, conn)
			return
		case "unsubscribe":
			s.unsubscribe(bus.Event, conn)
			return
		case "publish":
			s.publish(bus.Event, string(data))
		}
	}
	tcpServer.OnClose = func(conn *server.Connect) {
		//关闭连接就把这个连接订阅的所有事件全部删除
		for event, _ := range s.eventsSubscribe {
			s.unsubscribe(event, conn)
		}
	}
	err = tcpServer.Run()
	return
}

//某个连接订阅事件
func (s *Server) subscribe(event string, conn *server.Connect) {
	lock.Lock()
	defer lock.Unlock()
	_, ok := s.eventsSubscribe[event]
	if !ok {
		s.eventsSubscribe[event] = []*server.Connect{conn}
	} else if !isContain(conn, s.eventsSubscribe[event]) {
		s.eventsSubscribe[event] = append(s.eventsSubscribe[event], conn)
	}
	return
}

//取消某个连接的订阅事件
func (s *Server) unsubscribe(event string, conn *server.Connect) {
	lock.Lock()
	defer lock.Unlock()
	sl, ok := s.eventsSubscribe[event]
	if !ok {
		return
	}
	i := -1
	for k, v := range sl {
		if v == conn {
			i = k
			break
		}
	}
	if i < 0 {
		return
	}
	s.eventsSubscribe[event] = append(sl[:i], sl[i+1:]...)
	return
}

//广播事件和数据
func (s *Server) publish(event string, data string) {
	lock.Lock()
	defer lock.Unlock()
	sl, ok := s.eventsSubscribe[event]
	if !ok {
		return
	}
	for _, v := range sl {
		(*v).Write(s.protocol.EncodeOnePkg(data))
	}
}

func isContain(ele *server.Connect, sl []*server.Connect) bool {
	for _, v := range sl {
		if v == ele {
			return true
		}
	}
	return false
}
