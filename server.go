package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip string
	Port int
	//在线用户表
	OnlineMap map[string]*User
	//OnlineMap读写锁
	mapLock sync.RWMutex
	//广播消息管道
	Message  chan string

}

// NewServer 创建一个server的接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip: ip,
		Port: port,
		OnlineMap: make(map[string]*User),
		Message: make(chan string),
	}

	return server
}
// 监听 message 一旦有消息就广播
func (t *Server) listenMessage()  {

	for  {
		msg := <- t.Message
		t.mapLock.RLock()
		for _,cli := range t.OnlineMap{
			cli.C <- msg
		}

		t.mapLock.RUnlock()
	}
}
// BroadCast 广播消息
func (t *Server) BroadCast(user *User, msg string)  {
	sendMsg := "[" + user.Addr + "]" +user.Name + ":" + msg

	t.Message <- sendMsg
}

func (t *Server) Handler (conn net.Conn){
	//fmt.Println("连接建立成功")
	user := NewUser(conn,t)

	user.Online()
	//接受客户端发送的消息

	isLive := make(chan bool)
	go func() {
		buf := make([]byte, 4096)
		for{
			n,err := conn.Read(buf)

			if n == 0{
				// n=0代表下线？？
				user.Offline()
				return
			}
			 if err != nil && err != io.EOF{

			 	fmt.Println("conn read err: ", err)
			 	return
			 }

			 msg := string(buf[:n-1])
			 user.doMessage(msg)
			 isLive <- true


		}

	}()

	//当前handler先阻塞避免执行完上面操作后死亡
	//实现超时踢出功能
	for{
		select {
		/*
			select-case中 case的顺序是随机的？
			为什么可以刷新计时器？
			情况1：select阻塞直到 case1触发，这种情况属于超时
			情况2：select阻塞的过程中 case2触发，select成功运行，进到下一个for select，此时就会执行 time.After,相当于刷新计时器。
		 */

		case <- time.After(time.Second*600):
			//case1
			user.sendMsg("你被踢了")
			//销毁资源
			close(user.C)
			//关闭连接
			conn.Close()
			//退出
			return


		case <- isLive:
			//case2

		}
	}

}

// Start 启动服务器的接口
func (t *Server) Start()  {
	// socket listen
	listener,err := net.Listen("tcp", fmt.Sprintf("%s:%d", t.Ip, t.Port))

	if err != nil{
		fmt.Println("net.Listener err:", err)
	}
	// close listen socket
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			fmt.Println("listener.Close err:", err)
		}
	}(listener)
	//启动监听message goroutine
	 go t.listenMessage()
	for {
		//accept 会阻塞直到客户连接进来
		conn ,err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}
		// do handler goroutine
		go t.Handler(conn)

	}
}

