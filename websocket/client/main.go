package main

import (
	"code.google.com/p/go.net/websocket"
	"math/rand"
	"time"
	"flag"
	"fmt"
)

var message []byte = []byte("hello, world!\n")


var counter int
func run(i int) {
	origin := "http://" + host +"/"
	url := "ws://" + host + ":80/echo"
	var ws *websocket.Conn
	var err error
	for {
		ws, err = websocket.Dial(url, "", origin)
		if err != nil {
			d := -1.;
			for d < 0.1 {
				d = rand.NormFloat64() + delay
			}
		} else {
			break
		}
	}
	counter += 1
	defer func() {
		counter -= 1
	}()


	for {
		d := -1.;
		for d < 0.1 {
			d = rand.NormFloat64() + delay
		}

		time.Sleep(time.Duration(d) * time.Second)

		t1 := time.Now()
		if _, err = ws.Write(message); err != nil {
			break
		}

		var msg = make([]byte, 512)
		if _, err = ws.Read(msg); err != nil {
			break
		}

		td := time.Since(t1)
		fmt.Printf("%6d %5.0fms\n", i, td.Seconds() * 1000)
	}
	fmt.Printf("err: %s\n", err.Error())
}

func runrun(i int) {
	for {
		t1 := time.Now()
		run(i)
		td := time.Since(t1)
		fmt.Printf("[!] Broken after %fs\n", td.Seconds())
		time.Sleep(10)
	}
}


func main() {
	flag.Parse()

	for i := 0; i < concurrency; i++ {
		go runrun(i)
	}

	for {
		time.Sleep(5 * time.Second)
		fmt.Printf("connections %d\n", counter)
	}
}

var concurrency int
var host string
var delay float64

func init() {
	flag.IntVar(&concurrency, "c", 1, "concurrency")
	flag.StringVar(&host, "host", "cf.popcount.org", "host")
	flag.Float64Var(&delay, "delay", 10, "delay")
}
