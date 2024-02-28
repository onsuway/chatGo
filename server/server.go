package server

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int

	// 在线用户的列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	// 消息广播的channel
	Message chan string
}

// NewServer 创建一个新的Server实例
func NewServer(ip string, port int) *Server {
	return &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
}

// ListenMessage 监听广播消息channel的goroutine，一旦有消息就发送给全部在线的用户
func (this *Server) ListenMessage() {
	for {
		msg := <-this.Message

		// 将msg发送给全部在线用户
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.CurChan <- msg
		}
		this.mapLock.Unlock()
	}

}

// BroadCast 广播消息的方法
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendMsg
}

// Handler 处理业务的方法
func (this *Server) Handler(conn net.Conn) {
	// 当前链接的业务
	//fmt.Println("连接建立成功")
	user := NewUser(conn, this)

	user.Online()

	// 监听用户是否活跃的channel
	isLive := make(chan bool)

	// 接收客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("conn.Read err:", err)
				return
			}
			// 提取用户消息（去除'\n'）
			msg := string(buf[:n-1])

			// 用户针对msg进行消息处理
			user.DoMessage(msg)

			// 用户的任意消息，代表当前用户是活跃的
			isLive <- true
		}
	}()

	// 当前handler阻塞
	for {
		select {
		case <-isLive:
			// 当前用户是活跃的，应该重置定时器
			// 不做任何事情，为了激活select，更新下面的定时器
		case <-time.After(60 * time.Second):
			// 超时
			// 将当前用户强制关闭
			user.SendMsg("Inactive, forced offline!\n")
			// 销毁资源
			close(user.CurChan)
			// 关闭连接
			err := conn.Close()
			if err != nil {
				fmt.Println("conn.Close err:", err)
				return
			}
			// 退出当前handler
			return
		}
	}

}

// Start 启动服务器
func (this *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}
	// close listen socket
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			fmt.Println("listener.Close err:", err)
		}
	}(listener)

	// 启动监听Message的goroutine
	go this.ListenMessage()

	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener.Accept err:", err)
			continue
		}
		// do handler
		go this.Handler(conn)
	}
}
