package main

import (
	"fmt"
	"log"
	"net/http"

	"goservice/api"
	"goservice/nacos"
)

// === 新增：健康检查处理器 ===
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	// 必须返回 200 OK 状态码
	w.WriteHeader(http.StatusOK)
	// 返回简单的 UP 文本
	w.Write([]byte("Service is UP"))
}

func main() {
	// 注册服务到 Nacos
	nacos.RegisterToNacos("goservice", "127.0.0.1", 8080)

	// HTTP 路由
	http.HandleFunc("/api/echo", api.EchoHandler)

	// === 关键：注册根路径处理器，用于 Nacos 健康检查 ===
	http.HandleFunc("/", HealthCheckHandler)

	fmt.Println("Go service running on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
