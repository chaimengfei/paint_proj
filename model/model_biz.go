package model

// ProductSimple simple格式
type ProductSimple struct {
	ID           int64   `json:"id"`
	Name         string  `json:"name"`
	SellerPrice  float64 `json:"seller_price"`
	CategoryID   int64   `json:"category_id"`
	CategoryName string  `json:"category_name"`
	Image        string  `json:"image"`
	Unit         string  `json:"unit"`
	Remark       string  `json:"remark"`
}
type ProductListResponse struct {
	Categories []Category                `json:"categories"`
	Products   map[int64][]ProductSimple `json:"products"`
}
type CartWithProduct struct {
	Cart

	ProductName        string  `json:"product_name"`
	ProductImage       string  `json:"product_image"`
	ProductSellerPrice float64 `json:"product_seller_price"`
	ProductUnit        string  `json:"product_unit"`
}

// 订单类的业务数据
type CheckoutOrderRequest struct {
	UserID      int64
	CartIDs     []int64       // 购物车ID列表
	BuyNowItems []*BuyNowItem // 立即购买商品
	AddressID   int64         // 收货地址ID
	CouponID    int64         // 优惠券ID
	Note        string        // 订单备注
}

type BuyNowItem struct {
	ProductID int64
	Quantity  int
}

type OrderListRequest struct {
	UserID   int64
	Status   int32
	Page     int32
	PageSize int32
}

type OrderPaidCallbackRequest struct {
	OrderNo       string
	PaymentNo     string
	PaymentType   int32
	PaymentTime   int64
	PaymentAmount float64
}

type CheckoutResponse struct {
	OrderItems    []*OrderItem `json:"order_items"`
	OrderNo       string       `json:"order_no"`
	TotalAmount   float64      `json:"total_amount"`
	ShippingFee   float64      `json:"shipping_fee"`
	PaymentAmount float64      `json:"payment_amount"`
}
