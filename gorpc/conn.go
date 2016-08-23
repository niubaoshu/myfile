package gorpc

import (
	"errors"
	"gorpc/utils"
	"io"
	//"log"
	"net"
	//"time"
)

type Conn struct {
	exitChan chan struct{}
	conn     *net.TCPConn
}

func NewConn(ch chan struct{}, c *net.TCPConn) *Conn {
	return &Conn{
		exitChan: ch,
		conn:     c,
	}
}

func (c *Conn) Produce() (interface{}, error) {

	select {
	case <-c.exitChan:
		return nil, errors.New("Server is exited")
	default:
	}

	l := make([]byte, 2)
	//[]byte{0x00, 0x00} 为心跳包，收到该包重启设置超时时间，并取下一个包
	// for l[0] == 0x00 && l[1] == 0x00 {
	// 	//超过10秒没有收到包超时，关闭连接

	// 	c.conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	if _, err := io.ReadFull(c.conn, l); err != nil {
		c.conn.Close()
		return l, err
	}
	// }

	length := (int(l[0]) << 8) | int(l[1])

	packet := utils.GetNByte(length)
	if _, err := io.ReadFull(c.conn, packet); err != nil {
		c.conn.Close()
		return packet, err
	}
	//log.Println("收到", c.conn.RemoteAddr(), "  的数据", l, packet)
	return packet, nil
}
