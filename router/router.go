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
	// 1.2 初始化配置文件，放在全局的Cfg
	configs.InitConfig()

	// 2.添加CORS中间件
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Types")
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
	userRepo := repository.NewUserRepository(db)
	addressRepo := repository.NewAddressRepository(db)
	stockRepo := repository.NewStockRepository(db)

	// 4.初始化服务层
	cartService := service.NewCartService(cartRepo, productRepo)
	productService := service.NewProductService(productRepo)
	orderService := service.NewOrderService(orderRepo, cartRepo, productRepo, addressRepo, stockRepo)
	payService := service.NewPayService(orderRepo, cartRepo, productRepo)
	userService := service.NewUserService(userRepo)
	addressService := service.NewAddressService(addressRepo)
	stockService := service.NewStockService(stockRepo, productRepo)

	// 5. 初始化控制器
	cartController := controller.NewCartController(cartService)
	productController := controller.NewProductController(productService)
	orderController := controller.NewOrderController(orderService)
	payController := controller.NewPayController(payService)
	userController := controller.NewUserController(userService)
	addressController := controller.NewAddressController(addressService)
	stockController := controller.NewStockController(stockService)

	// API路由 供微信小程序用
	api := r.Group("/api")
	{
		userGroup := api.Group("/user")
		{
			userGroup.POST("/login", userController.Login) // 首次登陆注册user_id并获取token
			userGroup.POST("/update/info", auth.AuthMiddleware(), userController.UpdateUserInfo)
		}
		productGroup := api.Group("/product")
		{
			productGroup.GET("/list", productController.GetProductList)
		}
		addressGroup := api.Group("/address", auth.AuthMiddleware())
		{
			addressGroup.GET("/list", addressController.GetAddressList)
			addressGroup.POST("/create", addressController.CreateAddress)
			addressGroup.POST("/set_default/:id", addressController.SetDefultAddress)
			addressGroup.POST("/update", addressController.UpdateAddress)
			addressGroup.DELETE("/delete/:id", addressController.DeleteAddress)
		}
		cartGroup := api.Group("/cart", auth.AuthMiddleware())
		{
			cartGroup.GET("/list", cartController.GetCartList)
			cartGroup.POST("/add", cartController.AddToCart)
			cartGroup.POST("/update", cartController.UpdateCartItem)
			cartGroup.DELETE("/delete/:id", cartController.DeleteCartItem)
		}
		orderGroup := api.Group("/order", auth.AuthMiddleware())
		{
			orderGroup.GET("/list", orderController.GetOrderList)
			orderGroup.GET("/detail", orderController.GetOrderDetail)
			orderGroup.DELETE("/delete/:id", orderController.DeleteOrder) // 前端未用到

			orderGroup.POST("/checkout", orderController.CheckoutOrder)
			orderGroup.POST("/cancel", orderController.CancelOrder)
		}
		payGroup := api.Group("/pay")
		{

			payGroup.POST("/data", auth.AuthMiddleware(), payController.PaymentData)
			payGroup.POST("/callback", payController.PaymentCallback)
		}
	}
	// Admin路由 供Web后台管理系统
	admin := r.Group("/admin")
	{
		productGroup := admin.Group("/product")
		{
			productGroup.POST("/upload/image", productController.UploadImageForAdmin) // 阿里云OSS上传接口
			productGroup.GET("/list", productController.GetAdminProductList)
			productGroup.POST("/add", productController.AddProduct)
			productGroup.PUT("/edit/:id", productController.EditProduct)
			productGroup.DELETE("/del/:id", productController.DeleteProduct)
		}

		stockGroup := admin.Group("/stock")
		{
			stockGroup.POST("/batch/inbound", stockController.BatchInboundStock)      // 批量入库
			stockGroup.POST("/batch/outbound", stockController.BatchOutboundStock)    // 批量出库
			stockGroup.GET("/operations", stockController.GetStockOperations)         // 库存操作列表
			stockGroup.GET("/operation/:id", stockController.GetStockOperationDetail) // 库存操作详情
		}
	}

	return r
}
