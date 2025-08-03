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
	OrderItems    []*OrderItem `json:"order_items"`
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
