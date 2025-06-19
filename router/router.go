package router

import (
	"cmf/paint_proj/auth"
	"cmf/paint_proj/configs"
	"cmf/paint_proj/controller"
	"cmf/paint_proj/repository"
	"cmf/paint_proj/service"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// 1.初始化数据库
	db, err := configs.InitDB()
	if err != nil {
		panic("failed to connect database")
	}

	// 2.添加CORS中间件
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// 3.初始化仓储层
	cartRepo := repository.NewCartRepository(db)
	productRepo := repository.NewProductRepository(db)
	orderRepo := repository.NewOrderRepository(db)

	// 4.初始化服务层
	cartService := service.NewCartService(cartRepo, productRepo)
	productService := service.NewProductService(productRepo)
	orderService := service.NewOrderService(orderRepo, cartRepo, productRepo)

	// 5. 初始化控制器
	cartController := controller.NewCartController(cartService)
	productController := controller.NewProductController(productService)
	orderController := controller.NewOrderController(orderService)

	// API路由
	api := r.Group("/api")
	{
		productGroup := api.Group("/product")
		{
			productGroup.GET("/list", productController.GetProductList)
		}
		cartGroup := api.Group("/cart")
		{
			cartGroup.GET("/list", auth.AuthMiddleware(), cartController.GetCartList)
			cartGroup.POST("/add", auth.AuthMiddleware(), cartController.AddToCart)
			cartGroup.POST("/update", auth.AuthMiddleware(), cartController.UpdateCartItem)
			cartGroup.POST("/delete", auth.AuthMiddleware(), cartController.DeleteCartItem)
		}
		orderGroup := api.Group("/order")
		{
			orderGroup.GET("/list", auth.AuthMiddleware(), orderController.GetOrderList)
			orderGroup.GET("/detail", auth.AuthMiddleware(), orderController.GetOrderDetail)
			orderGroup.POST("/create", auth.AuthMiddleware(), orderController.CreateOrder)
			orderGroup.POST("/delete", auth.AuthMiddleware(), orderController.DeleteOrder)
			orderGroup.POST("/cancel", auth.AuthMiddleware(), orderController.CancelOrder)
			orderGroup.POST("/pay", auth.AuthMiddleware(), orderController.PayOrder)
			orderGroup.POST("/pay_callback", auth.AuthMiddleware(), orderController.PaymentCallback)
		}
	}

	return r
}
