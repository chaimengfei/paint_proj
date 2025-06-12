package main

import (
	"cmf/paint_proj/router"
	"log"
)

func main() {
	r := router.SetupRouter()
	// 启动服务

	if err := r.Run(":8009"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
