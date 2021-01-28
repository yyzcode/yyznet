package channel

import (
	"encoding/json"
	"fmt"
	"github.com/yyzcoder/yyznet/protocol"
	"net"
	"sync"
	"time"
)

type Client struct {
	ChannelAddr string
	ChannelPort int
	WaitRun sync.WaitGroup
	protocol    protocol.Transport
	status      int8
	conn        net.Conn
	subStatus   map[string]bool
	callback    map[string]func(data string)
}

//运行client
func (c *Client) Run() {
	c.WaitRun.Add(1)
	c.protocol = protocol.Frame{}//使用frame协议
	if c.ChannelPort > 65536 || c.ChannelPort < 1 {
		fmt.Println("channel port mast between 1 and 65535")
		c.WaitRun.Done()
	}
	var err error
	c.conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", c.ChannelAddr, c.ChannelPort))
	if err != nil {
		fmt.Println(err)
		c.WaitRun.Done()
	}
	c.status = 1
	c.WaitRun.Done()
	c.subAll()
	defer c.conn.Close()
	fmt.Println("channel server connection successful ", fmt.Sprintf("%s:%d", c.ChannelAddr, c.ChannelPort))
	bus := new(BusData)
	for {
		pkg, err := c.protocol.ReadOnePkg(&c.conn)
		if err != nil {
			//断线重连
			c.reconnect()
			continue
		}
		if err = json.Unmarshal(pkg, bus); err != nil {
			fmt.Println("channel数据格式错误:", err)
			continue
		}
		if fn, ok := c.callback[bus.Event]; ok {
			fn(bus.Data)
		}
	}
}

//订阅事件并设置回调函数
func (c *Client) On(event string, callback func(data string)) {
	if c.callback == nil {
		c.callback = make(map[string]func(data string), 0)
		c.subStatus = make(map[string]bool, 0)
	}
	c.callback[event] = callback
	if c.status != 1 {
		c.subStatus[event] = false
		return
	}
	if c.subStatus[event] == false{
		c.subscribes(event)
		c.subStatus[event] = true
	}
}

//取消事件订阅并删除回调函数
func (c *Client) Un(event string){
	bus := BusData{"unsubscribe", event, ""}
	jsonBytes, err := json.Marshal(bus)
	if err != nil {
		fmt.Println(err)
		return
	}
	if c.status != 1 {
		fmt.Println("隧道尚未连接成功")
		return
	}
	if _, err := c.conn.Write(c.protocol.EncodeOnePkg(string(jsonBytes))); err != nil {
		fmt.Println("取消订阅事件未成功", event)
		fmt.Println(err)
		return
	}
	delete(c.callback,event)
	delete(c.subStatus,event)
}

//发送订阅请求
func (c *Client) subscribes(event string) {
	bus := BusData{"subscribe", event, ""}
	jsonBytes, err := json.Marshal(bus)
	if err != nil {
		fmt.Println(err)
		return
	}
	if c.status != 1 {
		fmt.Println("隧道尚未连接成功")
		return
	}
	if _, err := c.conn.Write(c.protocol.EncodeOnePkg(string(jsonBytes))); err != nil {
		fmt.Println("订阅事件未成功", event)
		fmt.Println(err)
	}
}

//订阅所有已设置事件
//主要用于先设置订阅事件后运行客户端的情况，以及断线重连后的订阅
func (c *Client) subAll() {
	for event, status := range c.subStatus {
		if status == true {
			continue
		}
		c.subscribes(event)
	}
}

//广播事件和数据
func (c *Client) Publish(event string, data string) {
	if c.status != 1 {
		fmt.Println("隧道尚未连接成功")
		return
	}
	bus := BusData{
		InfoType: "publish",
		Event:    event,
		Data:     data,
	}
	jsonBytes, err := json.Marshal(bus)
	if err != nil {
		fmt.Println(err)
	}
	_, err = c.conn.Write(c.protocol.EncodeOnePkg(string(jsonBytes)))
	if err != nil {
		fmt.Println(err)
	}
}

//隧道重连
func (c *Client) reconnect() {
	var err error
	c.status = 2
RECONNECT:
	fmt.Println("正在重新连接隧道")
	time.Sleep(time.Second*5)//5s重连一次
	c.conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", c.ChannelAddr, c.ChannelPort))
	if err != nil {
		goto RECONNECT
	}
	//重连后重新订阅几个事件
	for event, _ := range c.subStatus {
		c.subStatus[event] = false
	}
	c.status = 1
	c.subAll()
}
