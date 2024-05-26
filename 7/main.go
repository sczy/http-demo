package main

import (
	"fmt"
	"log"
	"net/http"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func main() {
	// 创建一个简单的处理器函数
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %s", r.URL.Path[1:])
	})

	// 创建一个h2c服务器
	server := &http.Server{
		Addr:    ":8080",
		Handler: h2c.NewHandler(handler, &http2.Server{}),
	}

	log.Println("Starting h2c server on :8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("ListenAndServe: %v", err)
	}
}
