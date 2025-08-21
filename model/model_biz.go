package model

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
	OrderItems    []OrderItem  `json:"order_items"`
	OrderNo       string       `json:"order_no"`
	TotalAmount   Amount       `json:"total_amount"`
	ShippingFee   Amount       `json:"shipping_fee"`
	PaymentAmount Amount       `json:"payment_amount"`
	AddressData   *AddressInfo `json:"address_info"`
}

type LoginRequest struct {
	Code     string `json:"code"`
	Nickname string `json:"nickname"` // 小程序传来的昵称
	Avatar   string `json:"avatar"`   // 小程序传来的头像
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
	Name          string `json:"name" binding:"required"`        // 商品全名
	CategoryId    int64  `json:"category_id" binding:"required"` // 分类ID
	Image         string `json:"image" binding:"required"`       // 商品图片
	SellerPrice   Amount `json:"seller_price"`                   // 单价
	Cost          Amount `json:"cost"`                           // 成本价
	ShippingCost  Amount `json:"shipping_cost"`                  // 运费
	ProductCost   Amount `json:"product_cost"`                   // 货物成本
	Specification string `json:"specification"`                  // 规格（可选）
	Unit          string `json:"unit"`                           // 单位（可选）
	Remark        string `json:"remark"`                         // 备注（可选）
	IsOnShelf     int8   `json:"is_on_shelf"`                    // 是否上架(1:上架,0:下架)
}

// 库存操作类型常量
const (
	StockTypeInbound  = 1 // 入库
	StockTypeOutbound = 2 // 出库
	StockTypeReturn   = 3 // 退货
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
	Remark      string             `json:"remark"`                         // 备注
}

// 批量入库商品项
type BatchInboundItem struct {
	ProductID     int64  `json:"product_id" binding:"required"` // 商品ID
	Quantity      int    `json:"quantity" binding:"required"`   // 入库数量
	UnitPrice     Amount `json:"unit_price"`                    // 单价（可选）
	Remark        string `json:"remark"`                        // 备注（可选）
	ProductName   string `json:"product_name"`                  // 商品全名（自动补齐）
	Specification string `json:"specification"`                 // 规格（自动补齐）
	Unit          string `json:"unit"`                          // 单位（自动补齐）
	TotalPrice    Amount `json:"total_price"`                   // 总金额（自动计算）
}

// 批量出库请求结构体
type BatchOutboundRequest struct {
	Items       []BatchOutboundItem `json:"items" binding:"required"`        // 出库商品列表
	TotalAmount Amount              `json:"total_amount"`                    // 总金额（前端计算）
	UserName    string              `json:"user_name" binding:"required"`    // 用户名称
	UserID      int64               `json:"user_id" binding:"required"`      // 用户ID
	UserAccount string              `json:"user_account" binding:"required"` // 用户账号
	Operator    string              `json:"operator" binding:"required"`     // 操作人
	OperatorID  int64               `json:"operator_id" binding:"required"`  // 操作人ID
	Remark      string              `json:"remark"`                          // 备注
}

// 批量出库商品项
type BatchOutboundItem struct {
	ProductID     int64  `json:"product_id" binding:"required"` // 商品ID
	Quantity      int    `json:"quantity" binding:"required"`   // 出库数量
	UnitPrice     Amount `json:"unit_price" binding:"required"` // 单价
	ProductName   string `json:"product_name"`                  // 商品全名（自动补齐）
	Unit          string `json:"unit"`                          // 单位（自动补齐）
	Specification string `json:"specification"`                 // 规格（自动补齐）
	TotalPrice    Amount `json:"total_price"`                   // 总金额（自动计算）
	Remark        string `json:"remark"`                        // 备注（可选）
}
