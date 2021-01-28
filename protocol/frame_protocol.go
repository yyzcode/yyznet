package protocol

import (
	"encoding/binary"
	"fmt"
	"net"
)

type Frame struct {
}

//读取一个协议数据包
func (f Frame) ReadOnePkg(conn *net.Conn) (pkg []byte, err error) {
	//先读包头长度
	pkgLen, err := f.readPkgLen(conn)
	if err != nil {
		return
	}
	//读取包头所描述的长度
	pkg, err = f.readBytesUntilLen(conn, pkgLen)
	if err != nil {
		return
	}
	//循环读完数据后，发现数据包没有包头那么长，丢弃包
	if len(pkg) != pkgLen {
		err = fmt.Errorf("error package length")
		fmt.Println("error package length")
		pkg = pkg[0:0]
	}
	return
}

//从连接中读满长度length数据
func (f Frame) readBytesUntilLen(conn *net.Conn, length int) (b []byte, err error) {
	b = make([]byte, 0, length)
	//循环读取数据，直到达到length长度，网络不好时可能一次读不满，所以要循环读
	for x := length; x > 0; {
		var buf []byte
		buf, err = f.readBytesFromConn(conn, x)
		b = append(b, buf...)
		x = length - len(b)
		if err != nil { //产生错误跳出循环
			break
		}
	}
	return
}

//尝试从连接中读取x长度的数据
func (f Frame) readBytesFromConn(conn *net.Conn, x int) (b []byte, err error) {
	//从连接里读取x长度的数据
	read := make([]byte, x, x)
	n, err := (*conn).Read(read[:])
	b = read[:n]
	return
}

//读取frame协议包头长度
func (f Frame) readPkgLen(conn *net.Conn) (pkgLen int, err error) {
	//读取包头长度
	b, err := f.readBytesUntilLen(conn, 4)
	if err != nil {
		return
	}
	pkgLen = f.parseBigEndian32(b)
	return
}

//将数据用协议打包
func (f Frame) EncodeOnePkg(data string) []byte {
	pkg := f.bigEndian32(int32(len(data)))
	pkg = append(pkg, []byte(data)...)
	return pkg
}

func (f Frame) bigEndian32(i int32) []byte { // 大端序
	var b = make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(i)) //大端序模式
	return b
}

func (f Frame) parseBigEndian32(bytes []byte) int {
	return int(binary.BigEndian.Uint32(bytes[:]))
}
