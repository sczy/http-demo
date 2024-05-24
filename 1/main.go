package main

import (
	"fmt"
	"net"
)

func accept_request_thread(conn net.Conn) {
	defer conn.Close()
	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Printf("客户端退出 error=%v\n", err)
			return
		}
		fmt.Printf("接受消息 %s\n", string(buf[:n]))
	}
}

func main() {
	listen, err := net.Listen("tcp", "127.0.0.1:8000")
	if err != nil {
		fmt.Printf("监听失败 error=%v", err)
		return
	}
	fmt.Println("套接字创建成功，开始监听「127.0.0.1:8000」...")
	defer listen.Close()
	for {
		fmt.Println("等待客户端连接...")
		conn, err := listen.Accept()
		if err != nil {
			fmt.Printf("客户端连接失败 error=%v", err)
			continue
		} else {
			fmt.Println("通信套接字创建成功，开始接收消息...")
		}
		go accept_request_thread(conn)
	}
}
