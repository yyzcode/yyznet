package protocol

import (
	"bytes"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
)

type HttpParse struct {
	buffer []byte
}

func (h *HttpParse) ReadOnePkg(conn *net.Conn) (pkg []byte, err error) {
	httpHeader, err := h.readHttpHeader(conn)
	if err != nil {
		return
	}
	bodySize := h.getBodySize(httpHeader)
	var httpBody []byte
	if bodySize == 0 {
		pkg = httpHeader
	} else if bodySize == -1 { //chunk模式
		httpBody, err = h.readBytesUntilEndChunk(conn)
		if err != nil {
			fmt.Println("err package:", err)
			return
		}
	} else { //content-length模式
		httpBody, err = h.readBytesUntilLen(conn, bodySize)
		if err != nil {
			fmt.Println("err package:", err)
			return
		}
	}
	pkg = append(httpHeader, httpBody...)
	return
}

func (h *HttpParse) readHttpHeader(conn *net.Conn) (header []byte, err error) {
	var b []byte
	for {
		b, err = h.readBytesFromConn(conn, 512)
		h.buffer = append(h.buffer, b...)
		if index := bytes.Index(h.buffer, []byte("\r\n\r\n")); index > 0 {
			header = h.buffer[:index+4]
			h.buffer = h.buffer[(index + 4):]
			return
		}
		if err != nil {
			return
		}
	}
}

func (h *HttpParse) getBodySize(header []byte) (bodySize int) {
	headerStr := strings.ToLower(string(header))
	if strings.Contains(headerStr, "transfer-encoding: chunked") {
		return -1
	}
	reg := regexp.MustCompile(`content-length: [\d]+`)
	matches := reg.FindAllString(headerStr, -1)
	if len(matches) == 0 {
		bodySize = 0
		return
	}
	temp := strings.Split(matches[0], " ")
	bodySize, _ = strconv.Atoi(strings.Trim(temp[1], " "))
	return
}

func (h *HttpParse) readBytesUntilLen(conn *net.Conn, length int) (b []byte, err error) {
	b = make([]byte, 0, length)
	b = h.buffer
	x := length - len(b)
	//循环读取数据，直到达到length长度，网络不好时可能一次读不满，所以要循环读
	for x > 0 {
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

func (h *HttpParse) readBytesUntilEndChunk(conn *net.Conn) (b []byte, err error) {
	b = make([]byte, 0, 0)
	b = h.buffer
	if len(b) > 5 && string(b[(len(b)-5):]) == "0\r\n\r\n" {
		return
	}
	//循环读取数据，直到达到数据末尾刚好是 0\r\n\r\n
	for {
		var buf []byte
		buf, err = h.readBytesFromConn(conn, 512)
		b = append(b, buf...)
		if string(b[(len(b)-5):]) == "0\r\n\r\n" {
			break
		}
		if err != nil { //产生错误跳出循环
			break
		}
	}
	return
}

//尝试从连接中读取x长度的数据
func (h *HttpParse) readBytesFromConn(conn *net.Conn, x int) (b []byte, err error) {
	//从连接里读取x长度的数据
	read := make([]byte, x, x)
	n, err := (*conn).Read(read[:])
	b = read[:n]
	return
}
