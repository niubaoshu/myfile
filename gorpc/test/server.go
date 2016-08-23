package main

import (
	"fmt"
	"gorpc"
	"time"
)

func main() {
	gorpc.Linsten(plus)
	gorpc.Linsten(sub)
	gorpc.Linsten(printMsg)
	s := gorpc.NewServer(2345)
	s.Start()
}

func plus(a, b int) int {
	return a + b
}

func sub(a, b int) int {
	return a - b
}

func printMsg(msg string) {
	fmt.Println(msg)
}
func timeout(msg string) {
	time.Sleep(5 * time.Second)
	fmt.Println(msg)
}
func add(a ...int) int {
	var c = 0
	for i := 0; i < len(a); i++ {
		c += a[i]
	}
	return c
}
