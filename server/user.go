package server

import (
	"net"
	"strings"
)

type User struct {
	Name    string
	Addr    string
	CurChan chan string
	conn    net.Conn
	server  *Server
}

// NewUser 创建一个新的User实例
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:    userAddr,
		Addr:    userAddr,
		CurChan: make(chan string),
		conn:    conn,
		server:  server,
	}
	// 启动监听当前user channel消息的goroutine
	go user.ListenMessage()

	return user
}

// Online 用户上线
func (this *User) Online() {
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	this.server.BroadCast(this, "online!")
}

// Offline 用户下线
func (this *User) Offline() {
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()

	this.server.BroadCast(this, "offline!")
}

// SendMsg 给当前用户的客户端发送消息
func (this *User) SendMsg(msg string) {
	this.conn.Write([]byte(msg))
}

// DoMessage 用户处理消息的业务
func (this *User) DoMessage(msg string) {
	if msg == "who" {
		// 查询当前在线用户都有谁
		this.server.mapLock.Lock()
		this.conn.Write([]byte("online users:\n"))
		for _, user := range this.server.OnlineMap {
			onlineMsg := "\t[" + user.Addr + "]" + user.Name + "\n"
			this.conn.Write([]byte(onlineMsg))
		}
		this.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		// 消息格式：rename|张三
		newName := msg[7:]
		_, ok := this.server.OnlineMap[newName]
		if ok {
			this.conn.Write([]byte("current username is already in use!\n"))
		} else {
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap, this.Name)
			this.server.OnlineMap[newName] = this
			this.server.mapLock.Unlock()

			this.Name = newName
			this.SendMsg("rename success: " + this.Name + "\n")
		}
	} else if len(msg) > 1 && msg[0] == '@' {
		// 私聊消息格式：@张三|hello!

		// 1. 获取对方用户名
		// 2. 获取消息内容
		// 3. 根据用户名得到对方User对象
		// 4. 如果对方存在，发送消息
		idx := strings.Index(msg, "|")
		remoteName := msg[1:idx]
		if remoteName == "" {
			this.SendMsg("private message format is wrong! format:'@username|msg'\n")
			return
		}
		remoteUser, ok := this.server.OnlineMap[remoteName]
		if !ok {
			this.SendMsg("username does not exist!\n")
			return
		}
		content := msg[idx+1:]
		if content == "" {
			this.SendMsg("private message content cannot be empty!\n")
			return
		}
		remoteUser.SendMsg(this.Name + " say to you: " + content + "\n")
	} else {
		this.server.BroadCast(this, msg)
	}
}

// ListenMessage 监听当前User channel的方法，一旦有消息就发送给客户端
func (this *User) ListenMessage() {
	for {
		msg := <-this.CurChan
		// 将msg发送给当前用户的客户端（windows系统换行符\r\n linux用\n）
		this.conn.Write([]byte(msg + "\r\n"))
	}
}
