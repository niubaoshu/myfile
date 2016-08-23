package gorpc

import (
	"errors"
	"gorpc/gotiny"
	"gorpc/utils"
	"log"
	"net"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

type safeMap struct {
	m       map[uint64]chan []byte
	rwMutex sync.RWMutex
}

func (sm *safeMap) Set(key uint64, ch chan []byte) {
	sm.rwMutex.Lock()
	sm.m[key] = ch
	sm.rwMutex.Unlock()
}

func (sm *safeMap) Get(key uint64) (ch chan []byte, ok bool) {
	sm.rwMutex.RLock()
	defer sm.rwMutex.RUnlock()
	ch, ok = sm.m[key]
	return
}

func (sm *safeMap) Del(key uint64) {
	sm.rwMutex.Lock()
	delete(sm.m, key)
	sm.rwMutex.Unlock()
}

type Client struct {
	funcSum   uint64              //函数数量
	seq       map[uint64]*uint64  //序列id
	findCh    map[uint64]*safeMap //线程安全的map
	port      int                 //端口号
	waitGroup *sync.WaitGroup     //等呆退出
	conn      *net.TCPConn        // 连接
	funcs     []interface{}       //函数集
	exitChan  chan struct{}
}

func NewClient(port int, funcs []interface{}) *Client {
	funcSum := uint64(len(funcs))
	c := &Client{
		funcSum:   funcSum,
		port:      port,
		waitGroup: new(sync.WaitGroup),
		seq:       make(map[uint64]*uint64, funcSum),
		findCh:    make(map[uint64]*safeMap, funcSum),
		exitChan:  make(chan struct{}),
		funcs:     funcs,
	}

	for i := uint64(0); i < c.funcSum; i++ {
		c.seq[i] = new(uint64)
		c.findCh[i] = &safeMap{m: make(map[uint64]chan []byte)}
	}
	return c
}

func (c *Client) Start() {
	var err error
	c.conn, err = net.DialTCP("tcp", nil, &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: c.port})
	if err != nil {
		panic(err.Error())
	} else {
		log.Println("连接成功", c.conn.RemoteAddr())
	}
	go utils.NewConsume(&Conn{exitChan: c.exitChan, conn: c.conn}, &ClientHandler{conn: c.conn, c: c}, 100, 400, c.waitGroup).Start()
}

func (c *Client) Stop() {
	c.waitGroup.Wait()
}

//第一个是uint64类型，函数的id,后面跟原样的参数，再后面跟返回值的指针
func (c *Client) RemoteCall(fID uint64, para ...interface{}) error {
	c.waitGroup.Add(1)
	defer c.waitGroup.Done()

	s := atomic.AddUint64(c.seq[fID], 1)
	ch := make(chan []byte)
	c.findCh[fID].Set(s, ch) //设置通道，以便接受

	ft := reflect.TypeOf(c.funcs[fID])
	b := gotiny.EncodeWithPrefix(gotiny.EncodeWithPrefix([]byte{0x00, 0x00}, &fID, &s),
		para[:ft.NumIn()]...)
	if err := WritPacket(c.conn, b); err != nil {
		return err
	}
	// else {
	//	log.Println("向  ", c.conn.RemoteAddr(), "发送数据", b)
	//}
	select {
	case b := <-ch:
		c.findCh[fID].Del(s)
		d := gotiny.NewDecoder(b)
		d.Decodes(para[ft.NumIn():]...)
		return nil
	case <-time.Tick(5 * time.Second):
		return errors.New("the revert packet is timeout")
	}
}

type ClientHandler struct {
	conn *net.TCPConn
	c    *Client
}

func (ch *ClientHandler) Consume(i interface{}) {
	d := gotiny.NewDecoder(i.([]byte))
	c, _ := ch.c.findCh[d.DecUint()].Get(d.DecUint())
	c <- d.Bytes()
}
