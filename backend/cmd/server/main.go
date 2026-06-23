// Package main 启动入口：装配配置 / 日志 / DB / 路由，并启动 HTTP 服务。
package main

import (
	"log"

	"student-system/internal/boot"
)

func main() {
	if err := boot.Run(); err != nil {
		log.Fatalf("server boot failed: %v", err)
	}
}
