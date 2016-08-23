package gorpc

import (
	"gorpc/gotiny"
	"gorpc/utils"
	"log"
	"net"
	"reflect"
	"sync"
)

type server struct {
	exitChan  chan struct{}   // notify all goroutines to shutdown
	waitGroup *sync.WaitGroup // wait for all goroutines
	port      int
}

// Start starts service

func NewServer(port int) *server {
	return &server{
		exitChan:  make(chan struct{}),
		waitGroup: new(sync.WaitGroup),
		port:      port,
	}
}

func (s *server) Start() {
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: s.port})
	if err != nil {
		panic("监听端口失败")
	}
	defer listener.Close()
	log.Println("监听端口", s.port, "成功")
	for {
		select {
		case <-s.exitChan:
			return
		default:
		}
		//	listener.SetDeadline(<-time.Tick(time.Second))
		if conn, err := listener.AcceptTCP(); err == nil {
			log.Println("收到连接", conn.RemoteAddr())
			go func() {
				s.waitGroup.Add(1)
				utils.NewConsume(&Conn{s.exitChan, conn}, &serverHandler{conn: conn}, 10, 40, s.waitGroup).Start()
				s.waitGroup.Done()
			}()
		} else {
			log.Println("连接出错", err.Error())
		}
	}
}

// Stop stops service
func (s *server) Stop() {
	close(s.exitChan)
	s.waitGroup.Wait()
}

type serverHandler struct {
	conn *net.TCPConn
}

func (h *serverHandler) Consume(i interface{}) {
	d := gotiny.NewDecoder(i.([]byte))
	funcNum := d.DecUint()
	seq := d.DecUint()

	if funcNum > uint64(len(funcs)) {
		log.Println("没有要调用的函数")
		return
	}
	ft := reflect.TypeOf(funcs[funcNum])
	//log.Println(seq, funcNum)

	vs := make([]reflect.Value, ft.NumIn())
	for i := 0; i < ft.NumIn(); i++ {
		//log.Println(ft.NumIn(), i)
		vs[i] = d.DecodeByType(ft.In(i))
	}
	rvs := reflect.ValueOf(funcs[funcNum]).Call(vs)
	// for i := 0; i < len(rvs); i++ {
	// 	log.Println(rvs[i].Interface())
	// }
	p = gotiny.EncodeValuesWithPrefix(gotiny.EncodeValuesWithPrefix([]byte{0x00, 0x00},
		reflect.ValueOf(funcNum), reflect.ValueOf(seq)), rvs...)
	length := len(p) - 2
	p[0] = byte(length >> 8)
	p[1] = byte(length)
	//log.Println("向  ", c.RemoteAddr(), "发送数据", p)
	h.conn.Write(p)
}
