package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/zhxx123/websocket/service"
)

var (
	upgrader = websocket.Upgrader{
		// 允许跨域
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func wsHandler(w http.ResponseWriter, r *http.Request) {
	var (
		wsConn *websocket.Conn
		err    error
		conn   *service.Connection
		data   []byte
	)
	// 交换升级协议
	if wsConn, err = upgrader.Upgrade(w, r, nil); err != nil {
		fmt.Println("err", err.Error())
		return
	}
	if conn, err = service.InitConnection(wsConn); err != nil {
		goto ERR
	}
	fmt.Println("new connection...")
	// 心跳,每隔一秒
	go func() {
		for {
			if err := conn.WriteMessage([]byte("heartbeat")); err != nil {
				return
			}
			time.Sleep(2 * time.Second)
		}
	}()
	for {
		if data, err = conn.ReadMessage(); err != nil {
			goto ERR
		}
		fmt.Println(string(data))
		if err = conn.WriteMessage(data); err != nil {
			goto ERR
		}
	}
ERR:
	conn.Close()

}
func main() {
	http.HandleFunc("/ws", wsHandler)
	http.ListenAndServe("0.0.0.0:7001", nil)
}
