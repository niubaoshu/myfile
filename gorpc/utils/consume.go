package utils

import (
	"log"
	"sync"
	"time"
)

type consumer interface {
	Consume(i interface{})
}

type producer interface {
	Produce() (interface{}, error)
}

type consume struct {
	c         consumer
	p         producer
	ch        chan interface{}
	num       int
	elastic   int
	exitCh    chan struct{}
	exitPch   chan struct{}
	chCap     int
	waitGroup *sync.WaitGroup
}

func NewConsume(p producer, c consumer, elastic, chCap int, w *sync.WaitGroup) *consume {
	con := &consume{
		c:         c,
		p:         p,
		exitCh:    make(chan struct{}),
		exitPch:   make(chan struct{}, chCap),
		ch:        make(chan interface{}, chCap),
		elastic:   elastic,
		chCap:     chCap,
		waitGroup: w,
	}
	return con
}

func (c *consume) Start() {

	go c.singeProduce()
	l := 0

	c.num += c.elastic
	c.waitGroup.Add(c.elastic)
	for i := 0; i < c.elastic; i++ {
		go c.consume()
	}

	for {
		select {
		case <-c.exitCh:
			return
		default:
		}

		l = len(c.ch)
		if l > c.chCap/2 {
			c.num += l - c.chCap/2
			c.waitGroup.Add(l - c.chCap/2)
			for i := 0; i < l-c.chCap/2; i++ {
				go c.consume()
			}
		}
		if l > c.num+c.elastic {
			c.num += l - c.num
			c.waitGroup.Add(l - c.num)
			for i := 0; i < l-c.num; i++ {
				go c.consume()
			}
		}

		if l < c.num-c.elastic {
			c.closeNC(l - c.num)
		}
		time.Sleep(2 * time.Second)

		log.Println("现在有", len(c.ch), "个未处理，共有", c.num, "个线程在处理")
	}
}

func (c *consume) stop() {
	close(c.exitCh)
	close(c.ch)
}

func (c *consume) consume() {
	defer c.waitGroup.Done()
	for {
		select {
		case <-c.exitPch:
			return
		case p, b := <-c.ch:
			if b {
				c.c.Consume(p)
			} else {
				return
			}
		}
	}
}

func (c *consume) singeProduce() {
	for {
		select {
		case <-c.exitCh:
			return
		default:
			p, err := c.p.Produce()
			if err != nil {
				c.stop()
				return
			}
			c.ch <- p
		}
	}
}

func (c *consume) closeNC(n int) {
	c.num -= n
	var s struct{}
	for i := 0; i < n; i++ {
		c.exitPch <- s
	}
}
