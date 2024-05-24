package main

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
)

func unimplemented(conn net.Conn) {
	var buf string
	buf = "HTTP/1.0 501 Method Not Implemented\r\n"
	_, _ = conn.Write([]byte(buf))
	buf = "Server: httpd/0.1.0\r\n"
	_, _ = conn.Write([]byte(buf))
	buf = "Content-Type: text/html\r\n"
	_, _ = conn.Write([]byte(buf))
	buf = "\r\n"
	_, _ = conn.Write([]byte(buf))
	buf = "<HTML><HEAD><TITLE>Methord Not Implemented\r\n"
	_, _ = conn.Write([]byte(buf))
	buf = "</TITLE></HEAD>\r\n"
	_, _ = conn.Write([]byte(buf))
	buf = "<BODY><P>HTTP request method not supported.</P>\r\n"
	_, _ = conn.Write([]byte(buf))
	buf = "</BODY></HTML>\r\n"
	_, _ = conn.Write([]byte(buf))
}

func accept_request_thread(conn net.Conn) {
	defer conn.Close()
	var i int
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Printf("客户端退出 error=%v\n", err)
		return
	}

	fmt.Printf("接受消息 %s\n", string(buf[:n]))

	i = 0
	var method_bt strings.Builder
	for i < n && buf[i] != ' ' {
		method_bt.WriteByte(buf[i])
		i++
	}
	method := method_bt.String()
	if method != "GET" {
		unimplemented(conn)
		return
	}
	for i < n && buf[i] == ' ' {
		i++
	}

	var url_bt strings.Builder
	for i < n && buf[i] != ' ' {
		url_bt.WriteByte(buf[i])
		i++
	}
	url := url_bt.String()
	if method == "GET" {
		var path, query_string string
		j := strings.IndexAny(url, "?")
		if j != -1 {
			path = url[:j]
			if j+1 < len(url) {
				query_string = url[j+1:]
			}
		} else {
			path = url
		}
		fmt.Printf("path=%s, query_string=%s\n", path, query_string)
		resp := execute(path, query_string)
		fmt.Println("resp=%s\n", string(resp))
		header(conn, "application/json", len(resp))
		_, err = conn.Write(resp)
		if err != nil {
			fmt.Printf("conn.Write error=%v\n", err)
		}
	}
}

func header(conn net.Conn, content_type string, length int) {
	var buf string
	buf = "HTTP/1.0 200 OK\r\n"
	_, _ = conn.Write([]byte(buf))
	buf = "Server: httpd/0.1.0\r\n"
	_, _ = conn.Write([]byte(buf))
	buf = "Content-Type: " + content_type + "\r\n"
	_, _ = conn.Write([]byte(buf))
	buf = "Content-Length: " + fmt.Sprintf("%d", length) + "\r\n"
	_, _ = conn.Write([]byte(buf))
	buf = "Custom-Data: test" + "\r\n"
	_, _ = conn.Write([]byte(buf))
	buf = "\r\n"
	_, _ = conn.Write([]byte(buf))
}

func execute(path string, query_string string) []byte {
	query_param := make(map[string]string)
	parse_query_string(query_string, query_param)
	if path == "/" {
		camera_id := query_param["camera_id"]
		resp := make(map[string]interface{})
		resp["camera_id"] = camera_id
		resp["code"] = 200
		resp["msg"] = "ok"

		rs, err := json.Marshal(resp)
		if err != nil {
			fmt.Printf("json.Marshal error=%v\n", err)
		}
		return rs
	} else if "get_abc" == path {
		return []byte("get_abc")
	}
	return []byte("do't match")
}

func parse_query_string(query_string string, query_param map[string]string) {
	kvs := strings.Split(query_string, "&")
	if len(kvs) == 0 {
		return
	}
	for _, kv := range kvs {
		kv := strings.Split(kv, "=")
		if len(kv) == 2 {
			query_param[kv[0]] = kv[1]
		}
	}
}

func main() {
	listen, err := net.Listen("tcp", ":8000")
	if err != nil {
		fmt.Printf("net.Listen error=%v\n", err)
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
