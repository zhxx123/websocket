package service

import (
	"errors"
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

type Connection struct {
	wsConn    *websocket.Conn
	inChan    chan []byte
	outChan   chan []byte
	closeChan chan byte
	isClosed  bool
	mu        sync.Mutex
}

func InitConnection(wsConn *websocket.Conn) (conn *Connection, err error) {
	conn = &Connection{
		wsConn:    wsConn,
		inChan:    make(chan []byte, 1000),
		outChan:   make(chan []byte, 1000),
		closeChan: make(chan byte, 1),
	}
	go conn.readLoop()
	go conn.writeLoop()
	return
}

func (conn *Connection) Close() {
	// 线程安全的,可多次调用
	conn.wsConn.Close()
	conn.mu.Lock()
	if !conn.isClosed {
		close(conn.closeChan)
		conn.isClosed = true
		fmt.Println("connect is closed")
	}
	conn.mu.Unlock()
}

// 读消息
func (conn *Connection) ReadMessage() (data []byte, err error) {
	select {
	case data = <-conn.inChan:
	case <-conn.closeChan:
		err = errors.New("connect is closed")
	}
	return
}

// 写消息
func (conn *Connection) WriteMessage(data []byte) (err error) {
	select {
	case conn.outChan <- data:
	case <-conn.closeChan:
		err = errors.New("connect is closed")
	}
	return
}

// 一直读取消息
func (conn *Connection) readLoop() {
	var (
		data []byte
		err  error
	)
	for {
		if _, data, err = conn.wsConn.ReadMessage(); err != nil {
			goto ERR
		}
		// 放入缓存,满了会阻塞，等待 inchan 空闲位置
		select {
		case conn.inChan <- data:
		case <-conn.closeChan:
			goto ERR
		}
	}
ERR:
	conn.Close()
}

func (conn *Connection) writeLoop() {
	var (
		data []byte
		err  error
	)
	for {
		select {
		case data = <-conn.outChan:
		case <-conn.closeChan:
			goto ERR
		}

		if err = conn.wsConn.WriteMessage(websocket.TextMessage, data); err != nil {
			goto ERR
		}
	}
ERR:
	conn.Close()
}
