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
	shopRepo := repository.NewShopRepository(db)
	operatorRepo := repository.NewOperatorRepository(db)

	// 4.初始化服务层
	cartService := service.NewCartService(cartRepo, productRepo, userRepo)
	productService := service.NewProductService(productRepo)
	orderService := service.NewOrderService(orderRepo, cartRepo, productRepo, addressRepo, stockRepo, userRepo)
	payService := service.NewPayService(orderRepo, cartRepo, productRepo)
	userService := service.NewUserService(userRepo, shopRepo)
	addressService := service.NewAddressService(addressRepo)
	stockService := service.NewStockService(stockRepo, productRepo)
	shopService := service.NewShopService(shopRepo)
	operatorService := service.NewOperatorService(operatorRepo, shopRepo)

	// 5. 初始化控制器
	cartController := controller.NewCartController(cartService)
	productController := controller.NewProductController(productService, userService, shopService)
	orderController := controller.NewOrderController(orderService)
	payController := controller.NewPayController(payService)
	userController := controller.NewUserController(userService, shopService)
	addressController := controller.NewAddressController(addressService, shopService)
	stockController := controller.NewStockController(stockService, productService)
	shopController := controller.NewShopController(shopService)
	operatorController := controller.NewOperatorController(operatorService)

	// API路由 供微信小程序用
	api := r.Group("/api")
	{
		api.POST("/login", userController.Login)          // 首次登陆注册user_id并获取token（无需token验证）
		api.GET("/shop/list", shopController.GetShopList) // 获取店铺列表（无需token验证）
		userGroup := api.Group("/user")
		{
			userGroup.POST("/update/info", auth.AuthMiddleware(), userController.UpdateUserInfo)
			userGroup.POST("/bind-mobile", auth.AuthMiddleware(), userController.WechatBindMobile) // 绑定手机号
		}
		productGroup := api.Group("/product", auth.AuthMiddleware())
		{
			productGroup.GET("/list", productController.GetProductList)
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
		addressGroup := api.Group("/address", auth.AuthMiddleware())
		{
			addressGroup.GET("/list", addressController.GetAddressList)
			addressGroup.POST("/create", addressController.CreateAddress)
			addressGroup.POST("/set_default/:id", addressController.SetDefultAddress)
			addressGroup.POST("/update", addressController.UpdateAddress)
			addressGroup.DELETE("/delete/:id", addressController.DeleteAddress)
		}
	}
	// Admin路由 供Web后台管理系统
	admin := r.Group("/admin")
	{
		// 管理员登录（无需token验证）
		admin.POST("/login", operatorController.AdminLogin)

		// 店铺接口（无需token验证）
		admin.GET("/shop/list", shopController.GetShopList) // 获取店铺列表

		// 需要认证的管理接口
		adminAuth := admin.Group("", auth.AdminAuthMiddleware())
		{
			// 管理员管理（需要超级管理员权限）
			operatorGroup := adminAuth.Group("/operator")
			{
				operatorGroup.GET("/list", operatorController.GetOperatorList) // 获取管理员列表
				operatorGroup.GET("/:id", operatorController.GetOperatorByID)  // 根据ID获取管理员信息
			}

			productGroup := adminAuth.Group("/product")
			{
				productGroup.POST("/upload/image", productController.UploadImageForAdmin) // 阿里云OSS上传接口
				productGroup.GET("/list", productController.GetAdminProductList)
				productGroup.GET("/:id", productController.GetProductByID) // 根据ID获取商品信息
				productGroup.POST("/add", productController.AddProduct)
				productGroup.PUT("/edit/:id", productController.EditProduct)
				productGroup.DELETE("/del/:id", productController.DeleteProduct)

				productGroup.GET("/categories", productController.GetCategories)           // 获取所有分类
				productGroup.POST("/category/add", productController.AddCategory)          // 新增分类
				productGroup.PUT("/category/edit/:id", productController.EditCategory)     // 编辑分类
				productGroup.DELETE("/category/del/:id", productController.DeleteCategory) // 删除分类
			}

			stockGroup := adminAuth.Group("/stock")
			{
				stockGroup.POST("/batch/inbound", stockController.BatchInboundStock)             // 批量入库
				stockGroup.POST("/batch/outbound", stockController.BatchOutboundStock)           // 批量出库
				stockGroup.POST("/set/payment-status", stockController.SetOutboundPaymentStatus) // 更新出库单支付状态
				stockGroup.GET("/operations", stockController.GetStockOperations)                // 库存操作列表
				stockGroup.GET("/operation/:id", stockController.GetStockOperationDetail)        // 库存操作详情
				stockGroup.GET("/items", stockController.GetStockOperationItems)                 // 库存操作明细列表
				stockGroup.GET("/suppliers", stockController.GetSupplierList)                    // 获取供货商列表
			}

			userGroup := adminAuth.Group("/user")
			{
				userGroup.GET("/list", userController.AdminGetUserList)      // 获取用户列表
				userGroup.GET("/:id", userController.AdminGetUserByID)       // 根据ID获取用户信息
				userGroup.POST("/add", userController.AdminAddUser)          // 添加用户
				userGroup.PUT("/edit", userController.AdminEditUser)         // 编辑用户
				userGroup.DELETE("/del/:id", userController.AdminDeleteUser) // 删除用户
			}

			addressGroup := adminAuth.Group("/address")
			{
				addressGroup.GET("/list", addressController.AdminAddressList)         // 地址列表（支持用户ID或用户名搜索）
				addressGroup.POST("/add", addressController.AdminCreateAddress)       // 新增地址
				addressGroup.PUT("/edit", addressController.AdminUpdateAddress)       // 编辑地址
				addressGroup.DELETE("/del/:id", addressController.AdminDeleteAddress) // 删除地址
			}
		}
	}
	return r
}
