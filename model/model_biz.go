package model

import "time"

// ProductSimple simple格式
type ProductSimple struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	SellerPrice  Amount `json:"seller_price"`
	CategoryID   int64  `json:"category_id"`
	CategoryName string `json:"category_name"`
	Image        string `json:"image"`
	Unit         string `json:"unit"`
	Remark       string `json:"remark"`
}
type ProductListResponse struct {
	Categories []Category                `json:"categories"`
	Products   map[int64][]ProductSimple `json:"products"`
}
type CartWithProduct struct {
	Cart

	ProductName        string `json:"product_name"`
	ProductImage       string `json:"product_image"`
	ProductSellerPrice Amount `json:"product_seller_price"`
	ProductUnit        string `json:"product_unit"`
}

// 订单类的业务数据
type CheckoutOrderRequest struct {
	UserID      int64
	CartIDs     []int64       // 购物车ID列表
	BuyNowItems []*BuyNowItem // 立即购买商品
	AddressID   int64         // 收货地址ID
	CouponID    int64         // 优惠券ID
}

type BuyNowItem struct {
	ProductID int64
	Quantity  int
}

type ProductIdReq struct {
	ProductID int64 `json:"product_id" binding:"required"`
}
type UpdateCartItemReq struct {
	CartID   int64 `json:"cart_id" binding:"required"`
	Quantity int   `json:"quantity" binding:"required,min=1"`
}
type OrderCheckoutReq struct {
	CartIDs   []int64 `json:"cart_ids"`
	ProductID int64   `json:"product_id"`
	Quantity  int     `json:"quantity"`
	AddressID int64   `json:"address_id"`
	CouponID  int64   `json:"coupon_id"`
}
type OrderNoReq struct {
	OrderNo string `json:"order_no"` // 订单号
}
type PayCallbackReq struct {
	OrderNo       string `json:"order_no"`
	PaymentNo     string `json:"payment_no"`
	PaymentType   int    `json:"payment_type"`
	PaymentTime   int64  `json:"payment_time"`
	PaymentAmount Amount `json:"payment_amount"`
}

type BuildPaymentParam struct {
	Code    string `json:"code"`     // ，前端通过 wx.login() 获取临时 code，后端就可以使用这个 code 请求微信服务器获取 openid 和 session_key
	OrderNo string `json:"order_no"` // 订单号
	Total   Amount `json:"total"`    // 单位：分
}

type PaidCallbackData struct {
	OrderNo       string
	PaymentNo     string
	PaymentType   int32
	PaymentTime   int64
	PaymentAmount Amount
}
type OrderListRequest struct {
	UserID   int64
	Status   int32
	Page     int32
	PageSize int32
}

type CheckoutResponse struct {
	Items         []StockOperationItem `json:"items"`
	OrderNo       string               `json:"order_no"`
	TotalAmount   Amount               `json:"total_amount"`
	ShippingFee   Amount               `json:"shipping_fee"`
	PaymentAmount Amount               `json:"payment_amount"`
	AddressData   *AddressInfo         `json:"address_info"`
}

type LoginRequest struct {
	Code      string  `json:"code"`
	Nickname  string  `json:"nickname"`  // 小程序传来的昵称
	Avatar    string  `json:"avatar"`    // 小程序传来的头像
	Latitude  float64 `json:"latitude"`  // 纬度（可选）
	Longitude float64 `json:"longitude"` // 经度（可选）
}

type UpdateUserInfoRequest struct {
	Nickname string `json:"nickname"`
	Mobile   string `json:"mobile"`
}

type SetAddressDefaultReq struct {
	AddressID int64 `json:"address_id" binding:"required"`
	IsDefault bool  `json:"is_default"`
}

type AddressInfo struct {
	AddressID      int64  `json:"address_id"`
	RecipientName  string `json:"recipient_name"`
	RecipientPhone string `json:"recipient_phone"`
	Province       string `json:"province"`
	City           string `json:"city"`
	District       string `json:"district"`
	Detail         string `json:"detail"`
	IsDefault      *bool  `json:"is_default"`
}
type CreateAddressReq struct {
	Data AddressInfo `json:"data"`
}
type UpdateAddressReq struct {
	Data AddressInfo `json:"data"`
}

// 简化的商品请求结构体
type AddOrEditSimpleProductRequest struct {
	Name          string `json:"name" binding:"required"`         // 商品全名
	CategoryId    int64  `json:"category_id" binding:"required"`  // 分类ID
	Image         string `json:"image" binding:"required"`        // 商品图片
	SellerPrice   Amount `json:"seller_price" binding:"required"` // 售价
	Specification string `json:"specification"`                   // 规格（可选）
	Unit          string `json:"unit" binding:"required"`         // 单位 L/桶/卷
	IsOnShelf     int8   `json:"is_on_shelf" binding:"required"`  // 是否上架(1:上架,0:下架)
	Remark        string `json:"remark"`                          // 备注（可选）
	ShopID        int64  `json:"shop_id"`                         // 店铺ID（可选，从JWT token中获取）
}

// 编辑商品请求结构体
type EditProductRequest struct {
	Name          string `json:"name" binding:"required"`         // 商品全名
	Image         string `json:"image" binding:"required"`        // 商品图片
	SellerPrice   Amount `json:"seller_price" binding:"required"` // 售价
	Specification string `json:"specification"`                   // 规格（可选）
	IsOnShelf     int8   `json:"is_on_shelf" binding:"required"`  // 是否上架(1:上架,0:下架)
	Remark        string `json:"remark"`                          // 备注（可选）
	ShopID        int64  `json:"shop_id"`                         // 店铺ID（可选，从JWT token中获取）
}

// 分类管理请求结构体
type AddCategoryRequest struct {
	Name      string `json:"name" binding:"required"` // 分类名称
	SortOrder int64  `json:"sort_order"`              // 排序权重(数字越大越靠前)
	ShopID    int64  `json:"shop_id"`                 // 店铺ID
}

type EditCategoryRequest struct {
	Name      string `json:"name" binding:"required"` // 分类名称
	SortOrder int64  `json:"sort_order"`              // 排序权重(数字越大越靠前)
	ShopID    int64  `json:"shop_id"`                 // 店铺ID
}

// 库存操作类型常量
const (
	StockTypeInbound  = 1 // 入库
	StockTypeOutbound = 2 // 出库
	StockTypeReturn   = 3 // 退货
)

// 出库类型常量
const (
	OutboundTypeMiniProgram = 1 // 小程序购买
	OutboundTypeAdmin       = 2 // admin后台操作
)

// 库存操作请求结构体
type StockOperationRequest struct {
	ProductID int64  `json:"product_id" binding:"required"` // 商品ID
	Quantity  int    `json:"quantity" binding:"required"`   // 操作数量
	Remark    string `json:"remark"`                        // 备注
}

// 批量入库请求结构体
type BatchInboundRequest struct {
	Items       []BatchInboundItem `json:"items" binding:"required"`       // 入库商品列表
	TotalAmount Amount             `json:"total_amount"`                   // 总金额（前端计算）
	Operator    string             `json:"operator" binding:"required"`    // 操作人
	OperatorID  int64              `json:"operator_id" binding:"required"` // 操作人ID
	ShopID      int64              `json:"shop_id" binding:"required"`     // 店铺ID
	Supplier    string             `json:"supplier"`                       // 供货商
	Remark      string             `json:"remark"`                         // 备注
}

// 批量入库商品项
type BatchInboundItem struct {
	ProductID   int64  `json:"product_id" binding:"required"`   // 商品ID
	Quantity    int    `json:"quantity" binding:"required"`     // 入库数量
	ProductCost Amount `json:"product_cost" binding:"required"` // 货物成本（进价）
	TotalPrice  Amount `json:"total_price" binding:"required"`  // 单个商品总价
	Remark      string `json:"remark"`                          // 备注（可选）
}

// 批量出库请求结构体
type BatchOutboundRequest struct {
	Items       []BatchOutboundItem `json:"items" binding:"required"`       // 出库商品列表
	TotalAmount Amount              `json:"total_amount"`                   // 总金额（自动计算）
	UserName    string              `json:"user_name" binding:"required"`   // 用户名称
	UserID      int64               `json:"user_id" binding:"required"`     // 用户ID
	Operator    string              `json:"operator" binding:"required"`    // 操作人
	OperatorID  int64               `json:"operator_id" binding:"required"` // 操作人ID
	ShopID      int64               `json:"shop_id" binding:"required"`     // 店铺ID
	OperateTime *time.Time          `json:"operate_time"`                   // 操作时间（可选，如果传了则填充到created_at）
	Remark      string              `json:"remark"`                         // 备注
}

// 批量出库商品项
type BatchOutboundItem struct {
	ProductID  int64  `json:"product_id" binding:"required"` // 商品ID
	Quantity   int    `json:"quantity" binding:"required"`   // 出库数量
	UnitPrice  Amount `json:"unit_price" binding:"required"` // 卖价
	TotalPrice Amount `json:"total_price"`                   // 总金额（自动计算）
	Remark     string `json:"remark"`                        // 备注（可选）
}

// 更新出库单支付完成状态请求
type UpdateOutboundPaymentStatusRequest struct {
	OperationID         int64             `json:"operation_id" binding:"required"`          // 出库单ID
	PaymentFinishStatus PaymentStatusCode `json:"payment_finish_status" binding:"required"` // 支付完成状态(1:未支付,3:已支付)
	Operator            string            `json:"operator" binding:"required"`              // 操作人
	OperatorID          int64             `json:"operator_id" binding:"required"`           // 操作人ID
	ShopID              int64             `json:"shop_id"`                                  // 店铺ID
}

// 后台用户管理请求结构体
type AdminUserAddRequest struct {
	AdminDisplayName string `json:"admin_display_name" binding:"required"` // 后台显示的客户名称
	MobilePhone      string `json:"mobile_phone" binding:"required"`       // 手机号
	ShopID           int64  `json:"shop_id"`                               // 店铺ID（可选，不传则默认燕郊店）
	Remark           string `json:"remark"`                                // 备注
}

type AdminUserEditRequest struct {
	ID               int64  `json:"id" binding:"required"` // 用户ID
	AdminDisplayName string `json:"admin_display_name"`    // 后台显示的客户名称
	MobilePhone      string `json:"mobile_phone"`          // 手机号
	IsEnable         int8   `json:"is_enable"`             // 是否启用
	ShopID           int64  `json:"shop_id"`               // 店铺ID
	Remark           string `json:"remark"`                // 备注
}

type AdminUserSearchRequest struct {
	Keyword  string `json:"keyword"`   // 搜索关键词（手机号或姓名）
	Page     int    `json:"page"`      // 页码
	PageSize int    `json:"page_size"` // 每页大小
}

// 小程序用户绑定手机号请求
type WechatBindMobileRequest struct {
	MobilePhone string `json:"mobile_phone" binding:"required"` // 手机号
}

// AdminAddressInfo admin地址管理信息
type AdminAddressInfo struct {
	AddressID      int64  `json:"address_id"`
	UserID         int64  `json:"user_id"`
	UserName       string `json:"user_name"`
	RecipientName  string `json:"recipient_name"`
	RecipientPhone string `json:"recipient_phone"`
	Province       string `json:"province"`
	City           string `json:"city"`
	District       string `json:"district"`
	Detail         string `json:"detail"`
	IsDefault      bool   `json:"is_default"`
	CreatedAt      string `json:"created_at"`
}

// AdminAddressListRequest admin地址列表请求
type AdminAddressListRequest struct {
	UserID   int64  `json:"user_id" form:"user_id"`     // 用户ID（可选）
	UserName string `json:"user_name" form:"user_name"` // 用户名（可选）
	Page     int    `json:"page" form:"page"`           // 页码
	PageSize int    `json:"page_size" form:"page_size"` // 每页大小
}

// AdminCreateAddressRequest admin创建地址请求
type AdminCreateAddressRequest struct {
	UserID         int64  `json:"user_id" binding:"required"`         // 用户ID
	ShopID         int64  `json:"shop_id"`                            // 店铺ID
	RecipientName  string `json:"recipient_name" binding:"required"`  // 收货人姓名
	RecipientPhone string `json:"recipient_phone" binding:"required"` // 收货人电话
	Province       string `json:"province" binding:"required"`        // 省份
	City           string `json:"city" binding:"required"`            // 城市
	District       string `json:"district" binding:"required"`        // 区县
	Detail         string `json:"detail" binding:"required"`          // 详细地址
	IsDefault      bool   `json:"is_default"`                         // 是否默认地址
}

// AdminUpdateAddressRequest admin更新地址请求
type AdminUpdateAddressRequest struct {
	ID             int64  `json:"id" binding:"required"`              // 地址ID
	UserID         int64  `json:"user_id" binding:"required"`         // 用户ID
	ShopID         int64  `json:"shop_id"`                            // 店铺ID
	RecipientName  string `json:"recipient_name" binding:"required"`  // 收货人姓名
	RecipientPhone string `json:"recipient_phone" binding:"required"` // 收货人电话
	Province       string `json:"province" binding:"required"`        // 省份
	City           string `json:"city" binding:"required"`            // 城市
	District       string `json:"district" binding:"required"`        // 区县
	Detail         string `json:"detail" binding:"required"`          // 详细地址
	IsDefault      bool   `json:"is_default"`                         // 是否默认地址
}
