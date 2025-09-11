package main

import (
	"cmf/paint_proj/router"
	"log"
	"time"
)

func init() {
	// 设置应用时区为北京时间
	loc, _ := time.LoadLocation("Asia/Shanghai")
	time.Local = loc
}
func main() {
	r := router.SetupRouter()
	// 启动服务

	if err := r.Run(":8009"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
