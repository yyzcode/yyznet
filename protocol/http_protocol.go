package protocol

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
)

type Http struct {
	buffer []byte
}

func (h Http) ReadOnePkg(conn *net.Conn) (pkg []byte, err error) {
	httpHeader,err := h.readHttpHeader(conn)
	if err != nil{
		return
	}
	bodySize := h.getBodySize(httpHeader)
	if bodySize == 0{
		pkg = httpHeader
	}else{
		var httpBody []byte
		httpBody,err = h.readBytesUntilLen(conn,bodySize)
		if err != nil{
			fmt.Println("err package:",err)
			pkg = pkg[0:0]
			return
		}
		pkg = append(httpHeader,httpBody...)
	}
	return
}

func (h Http) readHttpHeader(conn *net.Conn) (header []byte, err error){
	var buf []byte
	var b []byte
	for{
		if b,err = h.readBytesUntilLen(conn,4);err != nil{
			return
		}
		buf = append(buf,b...)
		if string(b) == "\r\n\r\n"{
			break
		}
	}
	return
}

func (h Http) getBodySize(header []byte) (bodySize int){
	headerStr := strings.ToLower(string(header))
	reg := regexp.MustCompile(`content-length: [\d]+`)
	matches := reg.FindAllString(headerStr, -1)
	if len(matches) == 0{
		bodySize = 0
		return
	}
	temp := strings.Split(matches[0]," ")
	bodySize,_ = strconv.Atoi(strings.Trim(temp[1]," "))
	return
}

func (h Http) readBytesUntilLen(conn *net.Conn, length int) (b []byte, err error) {
	b = make([]byte, 0, length)
	//循环读取数据，直到达到length长度，网络不好时可能一次读不满，所以要循环读
	for x := length; x > 0; {
		var buf []byte
		buf, err = h.readBytesFromConn(conn, x)
		b = append(b, buf...)
		x = length - len(b)
		if err != nil { //产生错误跳出循环
			break
		}
	}
	return
}

//尝试从连接中读取x长度的数据
func (h Http) readBytesFromConn(conn *net.Conn, x int) (b []byte, err error) {
	//从连接里读取x长度的数据
	read := make([]byte, x, x)
	n, err := (*conn).Read(read[:])
	b = read[:n]
	return
}