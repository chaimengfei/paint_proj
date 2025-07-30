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
	userRepo := repository.NewUserRepository(db)
	addressRepo := repository.NewAddressRepository(db)

	// 4.初始化服务层
	cartService := service.NewCartService(cartRepo, productRepo)
	productService := service.NewProductService(productRepo)
	orderService := service.NewOrderService(orderRepo, cartRepo, productRepo, addressRepo)
	payService := service.NewPayService(orderRepo, cartRepo, productRepo)
	userService := service.NewUserService(userRepo)
	addressService := service.NewAddressService(addressRepo)

	// 5. 初始化控制器
	cartController := controller.NewCartController(cartService)
	productController := controller.NewProductController(productService)
	orderController := controller.NewOrderController(orderService)
	payController := controller.NewPayController(payService)
	userController := controller.NewUserController(userService)
	addressController := controller.NewAddressController(addressService)

	// API路由
	api := r.Group("/api")
	{
		userGroup := api.Group("/user")
		{
			userGroup.POST("/login", userController.Login)                // 首次登陆注册user_id并获取token
			userGroup.POST("/update/info", userController.UpdateUserInfo) // 首次登陆注册user_id并获取token
		}
		addressGroup := api.Group("/address")
		{
			addressGroup.GET("/list", auth.AuthMiddleware(), addressController.GetAddressList)
			addressGroup.POST("/create", auth.AuthMiddleware(), addressController.CreateAddress)
			addressGroup.POST("/set_default/:id", auth.AuthMiddleware(), addressController.SetDefultAddress)
			addressGroup.POST("/update", auth.AuthMiddleware(), addressController.UpdateAddress)
			addressGroup.DELETE("/delete/:id", auth.AuthMiddleware(), addressController.DeleteAddress)
		}
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
			orderGroup.DELETE("/delete", auth.AuthMiddleware(), orderController.DeleteOrder)

			orderGroup.POST("/checkout", auth.AuthMiddleware(), orderController.CheckoutOrder)
			orderGroup.POST("/cancel", auth.AuthMiddleware(), orderController.CancelOrder)
		}
		payGroup := api.Group("/pay")
		{

			payGroup.POST("/data", auth.AuthMiddleware(), payController.PaymentData)
			payGroup.POST("/callback", auth.AuthMiddleware(), payController.PaymentCallback)
		}
	}

	return r
}
