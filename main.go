package main

import server2 "chatGo/server"

func main() {
	server := server2.NewServer("127.0.0.1", 8888)
	server.Start()
}
