package main

import (
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s host:port", os.Args[0])
	}
	//获取命令行参数 socket地址
	server := os.Args[1]
	addr, err := net.ResolveTCPAddr("tcp4", server)
	checkError(err)

	//建立tcp连接
	conn, err := net.DialTCP("tcp4", nil, addr)
	checkError(err)

	//向服务端发送数据
	_, err = conn.Write([]byte("GET / HTTP/1.0\r\n\r\n"))
	checkError(err)
	//接收响应
	response, _ := io.ReadAll(conn)
	fmt.Println(string(response))
	os.Exit(0)
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
