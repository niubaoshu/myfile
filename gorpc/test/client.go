package main

import (
	"fmt"
	"gorpc"
	//"time"
)

var (
	funcs  []interface{}
	client *gorpc.Client
)

func main() {
	funcs = []interface{}{
		plus,
		sub,
		printMsg,
	}
	client = gorpc.NewClient(2345, funcs)
	client.Start()
	for i := 0; i < 100000; i++ {
		go func() {
			r, err := plus(1, 2)
			if err != nil {
				fmt.Println(err.Error())
			} else {
				if r != 3 {
					fmt.Println("错了。", r)
				}
			}
			r, err = sub(1, 2)
			if err != nil {
				fmt.Println(err.Error())
			} else {
				if r != -1 {
					fmt.Println("错了。", r)
				}
			}
			err = printMsg("hello,rpc")
			if err != nil {
				fmt.Println(err.Error())
			}
		}()
		//time.Sleep(time.Microsecond)
	}
	var a chan int
	<-a
}

func plus(a, b int) (c int, err error) {
	err = client.RemoteCall(uint64(0), &a, &b, &c)
	return
}
func sub(a, b int) (c int, err error) {
	err = client.RemoteCall(uint64(1), &a, &b, &c)
	return
}
func printMsg(msg string) (err error) {
	err = client.RemoteCall(uint64(2), &msg)
	return
}
func timeout(msg string) (err error) {
	err = client.RemoteCall(uint64(3), &msg)
	return
}

// func add(a ...int) (c int, err error) {
// 	err = client.RemoteCall(uint64(4), a..., &c)
// 	return
// }
