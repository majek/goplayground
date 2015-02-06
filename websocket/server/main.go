package main

import (
	"code.google.com/p/go.net/websocket"
	"net/http"
	"time"
	"fmt"
)

var counter int
func EchoServer(ws *websocket.Conn) {
	counter += 1
	defer func() {
		counter -= 1
	}()
	for {
		var msg = make([]byte, 512)
		if _, err := ws.Read(msg); err != nil {
			break
		}

		if _, err := ws.Write(msg); err != nil {
			break
		}
	}
	ws.Close()
}

func main() {
	http.Handle("/echo", websocket.Handler(EchoServer))
	fmt.Printf("Listening ws on 0.0.0.0:8080/echo\n")
	go func() {
		for {
			time.Sleep(5 * time.Second)
			fmt.Printf("connections %d\n", counter)
		}
	}()
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
