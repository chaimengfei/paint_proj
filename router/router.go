package router

import (
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

	// 4.初始化服务层
	cartService := service.NewCartService(cartRepo, productRepo)
	productService := service.NewProductService(productRepo)

	// 5. 初始化控制器
	cartController := controller.NewCartController(cartService)
	productController := controller.NewProductController(productService)

	// API路由
	api := r.Group("/api")
	{
		productGroup := api.Group("/product")
		{
			productGroup.GET("/list", productController.GetProductList)
		}
		cartGroup := api.Group("/cart")
		{
			cartGroup.POST("/add", cartController.AddToCart)
			cartGroup.GET("/list", cartController.GetCartList)
			cartGroup.POST("/update", cartController.UpdateCartItem)
			cartGroup.POST("/delete", cartController.DeleteCartItem)
		}
	}

	return r
}
