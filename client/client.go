package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	mode       int // 当前客户端选择的模式
}

var serverIp string
var serverPort int

// ./client -ip 127.0.0.1 -port 8888
func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器IP地址(默认是127.0.0.1)")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器端口(默认是8888)")
}

func NewClient(serverIp string, serverPort int) *Client {
	// 创建客户端对象
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		mode:       -1,
	}
	// 连接server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial err:", err)
		return nil
	}
	client.conn = conn

	return client
}

func (client *Client) menu() bool {
	var mode int
	fmt.Println("1. public chat mode")
	fmt.Println("2. private chat mode")
	fmt.Println("3. update username")
	fmt.Println("0. exit")
	fmt.Scanln(&mode)
	if mode >= 0 && mode <= 3 {
		client.mode = mode
		return true
	} else {
		fmt.Println(">>>>> error: illegal menu number! <<<<<")
		return false
	}
}

func (client *Client) UpdateName() bool {
	fmt.Print(">>>>> please input new username:")
	fmt.Scanln(&client.Name)
	sendMsg := "rename|" + client.Name + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("UpdateName conn.Write err:", err)
		return false
	}
	return true
}

func (client *Client) PublicChat() {
	var chatMsg string
	fmt.Println(">>>>> please input message, input 'exit' to exit <<<<<")
	fmt.Scanln(&chatMsg)
	for chatMsg != "exit" {
		if len(chatMsg) > 0 {
			sendMsg := chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("PublicChat conn.Write err:", err)
				break
			}
		}
		chatMsg = ""
		fmt.Println(">>>>> please input message, input 'exit' to exit <<<<<")
		fmt.Scanln(&chatMsg)
	}
}

// SelectUsers 显示当前在线用户
func (client *Client) SelectUsers() {
	sendMsg := "who\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("SelectUsers conn.Write err:", err)
		return
	}
}

func (client *Client) PrivateChat() {
	var remoteName string
	var chatMsg string

	client.SelectUsers()
	fmt.Println(">>>>> please input private chat username, input 'exit' to exit <<<<<")
	fmt.Scanln(&remoteName)
	for remoteName != "exit" {
		fmt.Println(">>>>> please input message, input 'exit' to exit <<<<<")
		fmt.Scanln(&chatMsg)
		for chatMsg != "exit" {
			if len(chatMsg) > 0 {
				sendMsg := "@" + remoteName + "|" + chatMsg + "\n\n"
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("PrivateChat conn.Write err:", err)
					break
				}
			}
			chatMsg = ""
			fmt.Println(">>>>> please input message, input 'exit' to exit <<<<<")
			fmt.Scanln(&chatMsg)
		}
		client.SelectUsers()
		fmt.Println(">>>>> please input private chat username, input 'exit' to exit <<<<<")
		fmt.Scanln(&remoteName)
	}
}

func (client *Client) Run() {
	for client.mode != 0 {
		for client.menu() != true {
		}
		// 根据不同的模式处理不同的业务
		switch client.mode {
		case 1:
			// public chat mode
			fmt.Println(">>>>> public chat mode <<<<<")
			client.PublicChat()
		case 2:
			// private chat mode
			fmt.Println(">>>>> private chat mode <<<<<")
			client.PrivateChat()
		case 3:
			// update username
			fmt.Println(">>>>> update username <<<<<")
			client.UpdateName()
		}
	}
	fmt.Println(">>>>> exit <<<<<")
}

// DealResponse 处理server的响应消息, 直接显示在客户端的终端
func (client *Client) DealResponse() {
	// 一旦client.conn有数据，就直接拷贝到stdout标准输出上，永久阻塞监听
	_, err := io.Copy(os.Stdout, client.conn)
	if err != nil {
		fmt.Println("DealResponse io.Copy err:", err)
		return
	}
}

func main() {
	// 命令行解析
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>> link server failed...")
		return
	}
	fmt.Println(">>>>> link server success...")

	// 单独开启一个goroutine处理server的响应消息
	go client.DealResponse()

	// 启动客户端的业务
	client.Run()
}
