package model

import (
	"encoding/json"
	"fmt"
	"math"
	"time"
)

// Product 油漆表
type Product struct {
	ID           int64  `json:"id" gorm:"id,primaryKey;autoIncrement" ` // 分类ID
	Name         string `json:"name" gorm:"name"`                       // 商品名
	SellerPrice  Amount `json:"seller_price" gorm:"seller_price"`       // 卖价
	Cost         Amount `json:"cost" gorm:"cost"`                       // 成本价(暂不用)
	ShippingCost Amount `json:"shipping_cost" gorm:"shipping_cost"`     // 运费(暂不用)
	ProductCost  Amount `json:"product_cost" gorm:"product_cost"`       // 货物成本(暂不用)
	CategoryId   int64  `json:"category_id" gorm:"category_id"`         // 分类id
	Stock        int    `json:"stock" gorm:"stock"`                     // 库存
	Image        string `json:"image" gorm:"image"`                     // 图片地址
	Unit         string `json:"unit" gorm:"unit"`                       // 单位 L/桶/套
	Remark       string `json:"remark" gorm:"remark"`                   // 备注
	IsOnShelf    int8   `json:"is_on_shelf" gorm:"is_on_shelf"`         // 是否上架(1:上架,0:下架)
}

// TableName 表名称
func (*Product) TableName() string {
	return "product"
}

// Category 商品分类表
type Category struct {
	ID        int64  `json:"id" gorm:"id,primaryKey;autoIncrement" ` // 分类ID
	Name      string `json:"name" gorm:"name"`                       // 分类名称
	SortOrder int64  `json:"sort_order" gorm:"sort_order"`           // 排序权重(数字越大越靠前)
}

// TableName 表名称
func (*Category) TableName() string {
	return "category"
}

type Cart struct {
	ID        int64      `gorm:"id,primaryKey;autoIncrement" json:"id" `
	UserID    int64      `gorm:"column:user_id" json:"user_id"`
	ProductID int64      `gorm:"column:product_id" json:"product_id"`
	Quantity  int        `gorm:"column:quantity" json:"quantity"`
	Selected  bool       `gorm:"column:selected" json:"selected"`
	CreatedAt *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (Cart) TableName() string {
	return "cart"
}

// OrderStatusCode 订单状态
type OrderStatusCode int8

// PaymentStatusCode 支付状态
type PaymentStatusCode int8

// PaymentTypeCode 支付方式(1:微信支付,2:支付宝,3:余额支付)
type PaymentTypeCode int8

const (
	OrderStatusPendingPayment OrderStatusCode = 1 // 待付款
	OrderStatusPaymentSuccess OrderStatusCode = 2 // 已付款(待发货)
	OrderStatusPendingReceipt OrderStatusCode = 3 // 待收货
	OrderStatusCancelled      OrderStatusCode = 4 // 已取消
	OrderStatusCompleted      OrderStatusCode = 5 // 已完成

	PaymentStatusUnpaid    PaymentStatusCode = 1 // 未支付
	PaymentStatusPaying    PaymentStatusCode = 2 // 支付中
	PaymentStatusPaid      PaymentStatusCode = 3 // 已支付
	PaymentStatusRefunding PaymentStatusCode = 4 // 退款中
	PaymentStatusRefunded  PaymentStatusCode = 5 // 已退款
	PaymentStatusFailed    PaymentStatusCode = 6 // 支付失败

	PaymentTypeWX      PaymentTypeCode = 1
	PaymentTypeZFB     PaymentTypeCode = 2
	PaymentTypeBalance PaymentTypeCode = 3

	//  操作人类型
	OperatorTypeUser   = 1 // 用户
	OperatorTypeSystem = 2 // 系统
	OperatorTypeAdmin  = 3 // 管理员
)

// Order 订单表
type Order struct {
	ID              int64             `json:"id" gorm:"id,primaryKey;autoIncrement"`    // 主键id
	OrderNo         string            `json:"order_no" gorm:"order_no"`                 // 订单编号
	UserId          int64             `json:"user_id" gorm:"user_id"`                   // 用户ID
	TotalAmount     Amount            `json:"total_amount" gorm:"total_amount"`         // 订单总金额
	PaymentAmount   Amount            `json:"payment_amount" gorm:"payment_amount"`     // 实付金额
	ShippingFee     Amount            `json:"shipping_fee" gorm:"shipping_fee"`         // 运费
	DiscountAmount  Amount            `json:"discount_amount" gorm:"discount_amount"`   // 优惠金额
	CouponAmount    Amount            `json:"coupon_amount" gorm:"coupon_amount"`       // 优惠券抵扣金额
	PaymentType     PaymentTypeCode   `json:"payment_type" gorm:"payment_type"`         // 支付方式(1:微信支付,2:支付宝,3:余额支付)
	PaymentTime     *time.Time        `json:"payment_time" gorm:"payment_time"`         // 支付时间
	PaymentStatus   PaymentStatusCode `json:"payment_status" gorm:"payment_status"`     // 支付状态(1:未支付,2:支付中,3:已支付,4:退款中,5:已退款,6:支付失败)
	OrderStatus     OrderStatusCode   `json:"order_status" gorm:"order_status"`         // 订单状态(1:待付款,2:待发货,3:待收货,4:已取消,5:已完成)
	ReceiverName    string            `json:"receiver_name" gorm:"receiver_name"`       // 收货人姓名
	ReceiverPhone   string            `json:"receiver_phone" gorm:"receiver_phone"`     // 收货人电话
	ReceiverAddress string            `json:"receiver_address" gorm:"receiver_address"` // 收货地址
	Note            string            `json:"note" gorm:"note"`                         // 订单备注
	CreatedAt       *time.Time        `json:"created_at" gorm:"created_at"`             // 创建时间
	UpdatedAt       *time.Time        `json:"updated_at" gorm:"updated_at"`             // 更新时间
	DeletedAt       *time.Time        `json:"deleted_at" gorm:"deleted_at"`             // 删除时间

	OrderItems []OrderItem `json:"order_items" gorm:"-"` // ✅ 不映射到数据库，纯业务使用
}

// TableName 表名称
func (*Order) TableName() string {
	return "order"
}

// OrderItem 订单商品表
type OrderItem struct {
	ID           int64      `json:"id" gorm:"id,primaryKey;autoIncrement"` // 主键id
	OrderId      int64      `json:"order_id" gorm:"order_id"`              // 订单ID
	OrderNo      string     `json:"order_no" gorm:"order_no"`              // 订单编号
	ProductId    int64      `json:"product_id" gorm:"product_id"`          // 商品ID
	ProductName  string     `json:"product_name" gorm:"product_name"`      // 商品名称
	ProductImage string     `json:"product_image" gorm:"product_image"`    // 商品图片
	ProductPrice Amount     `json:"product_price" gorm:"product_price"`    // 商品单价
	TotalPrice   Amount     `json:"total_price" gorm:"total_price"`        // 商品总价
	Quantity     int        `json:"quantity" gorm:"quantity"`              // 购买数量
	Unit         string     `json:"unit" gorm:"unit"`                      // 商品单位
	CreatedAt    *time.Time `json:"created_at" gorm:"created_at"`          // 创建时间
	UpdatedAt    *time.Time `json:"updated_at" gorm:"updated_at"`          // 更新时间
}

// TableName 表名称
func (*OrderItem) TableName() string {
	return "order_item"
}

// OrderLog 订单操作日志表
type OrderLog struct {
	ID           int64      `json:"id" gorm:"id,primaryKey;autoIncrement"` // 主键id
	OrderId      int64      `json:"order_id" gorm:"order_id"`              // 订单ID
	OrderNo      string     `json:"order_no" gorm:"order_no"`              // 订单编号
	Action       string     `json:"action" gorm:"action"`                  // 操作行为
	Operator     string     `json:"operator" gorm:"operator"`              // 操作人
	OperatorType int8       `json:"operator_type" gorm:"operator_type"`    // 操作人类型(1:用户,2:系统,3:管理员)
	Content      string     `json:"content" gorm:"content"`                // 操作内容
	CreatedAt    *time.Time `json:"created_at" gorm:"created_at"`          // 创建时间
}

// TableName 表名称
func (*OrderLog) TableName() string {
	return "order_log"
}

// Payment 支付记录表
type Payment struct {
	ID              int64      `json:"id" gorm:"id,primaryKey;autoIncrement"`    // 主键id
	UserId          int64      `json:"user_id" gorm:"user_id"`                   // 用户ID
	OrderId         int64      `json:"order_id" gorm:"order_id"`                 // 订单ID
	OrderNo         string     `json:"order_no" gorm:"order_no"`                 // 订单编号
	PaymentNo       string     `json:"payment_no" gorm:"payment_no"`             // 支付流水号
	PaymentType     int8       `json:"payment_type" gorm:"payment_type"`         // 支付方式(1:微信支付,2:支付宝,3:余额支付)
	PaymentAmount   Amount     `json:"payment_amount" gorm:"payment_amount"`     // 支付金额
	PaymentStatus   int8       `json:"payment_status" gorm:"payment_status"`     // 支付状态(1:未支付,2:支付中,3:已支付,4:退款中,5:已退款,6:支付失败)
	PaymentTime     *time.Time `json:"payment_time" gorm:"payment_time"`         // 支付时间
	CallbackTime    *time.Time `json:"callback_time" gorm:"callback_time"`       // 回调时间
	CallbackContent string     `json:"callback_content" gorm:"callback_content"` // 回调内容
	CreatedAt       *time.Time `json:"created_at" gorm:"created_at"`             // 创建时间
	UpdatedAt       *time.Time `json:"updated_at" gorm:"updated_at"`             // 更新时间
}

// TableName 表名称
func (*Payment) TableName() string {
	return "payment"
}

// User 小程序用户表
type User struct {
	ID          int64     `json:"id" gorm:"id"`                     // 用户ID
	Openid      string    `json:"openid" gorm:"openid"`             // 微信OpenID
	Nickname    string    `json:"nickname" gorm:"nickname"`         // 微信昵称
	Avatar      string    `json:"avatar" gorm:"avatar"`             // 头像
	MobilePhone string    `json:"mobile_phone" gorm:"mobile_phone"` // 手机号
	CreatedAt   time.Time `json:"created_at" gorm:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"updated_at"`
}

// TableName 表名称
func (*User) TableName() string {
	return "user"
}

type Amount int64

func (a Amount) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%.2f", float64(a)/100)), nil
}

func (a *Amount) UnmarshalJSON(data []byte) error {
	var f float64
	if err := json.Unmarshal(data, &f); err != nil {
		return err
	}
	*a = Amount(math.Round(f * 100)) // 四舍五入
	return nil
}

// Address undefined
type Address struct {
	ID             int64  `json:"id" gorm:"id"`
	UserId         int64  `json:"user_id" gorm:"user_id"`
	RecipientName  string `json:"recipient_name" gorm:"recipient_name"`
	RecipientPhone string `json:"recipient_phone" gorm:"recipient_phone"`
	Province       string `json:"province" gorm:"province"`
	City           string `json:"city" gorm:"city"`
	District       string `json:"district" gorm:"district"`
	Detail         string `json:"detail" gorm:"detail"`
	IsDefault      int8   `json:"is_default" gorm:"is_default"`
	IsDelete       int8   `json:"is_delete" gorm:"is_delete"`
}

// TableName 表名称
func (*Address) TableName() string {
	return "address"
}

// 库存操作主表
type StockOperation struct {
	ID           int64      `json:"id" gorm:"id,primaryKey;autoIncrement"` // 主键id
	OperationNo  string     `json:"operation_no" gorm:"operation_no"`      // 操作单号
	Type         int8       `json:"type" gorm:"type"`                      // 操作类型(1:入库,2:出库,3:退货)
	Operator     string     `json:"operator" gorm:"operator"`              // 操作人
	OperatorID   int64      `json:"operator_id" gorm:"operator_id"`        // 操作人ID
	OperatorType int8       `json:"operator_type" gorm:"operator_type"`    // 操作人类型(1:用户,2:系统,3:管理员)
	UserName     string     `json:"user_name" gorm:"user_name"`            // 用户名称(出库时)
	UserID       int64      `json:"user_id" gorm:"user_id"`                // 用户ID(出库时)
	UserAccount  string     `json:"user_account" gorm:"user_account"`      // 用户账号(出库时)
	PurchaseTime *time.Time `json:"purchase_time" gorm:"purchase_time"`    // 购买时间(出库时)
	Remark       string     `json:"remark" gorm:"remark"`                  // 备注
	TotalAmount  Amount     `json:"total_amount" gorm:"total_amount"`      // 总金额
	CreatedAt    *time.Time `json:"created_at" gorm:"created_at"`          // 创建时间
}

// TableName 表名称
func (*StockOperation) TableName() string {
	return "stock_operation"
}

// 库存操作子表
type StockOperationItem struct {
	ID          int64      `json:"id" gorm:"id,primaryKey;autoIncrement"` // 主键id
	OperationID int64      `json:"operation_id" gorm:"operation_id"`      // 操作主表ID
	ProductID   int64      `json:"product_id" gorm:"product_id"`          // 商品ID
	ProductName string     `json:"product_name" gorm:"product_name"`      // 商品名称
	Quantity    int        `json:"quantity" gorm:"quantity"`              // 操作数量
	UnitPrice   Amount     `json:"unit_price" gorm:"unit_price"`          // 单价
	TotalPrice  Amount     `json:"total_price" gorm:"total_price"`        // 总价
	BeforeStock int        `json:"before_stock" gorm:"before_stock"`      // 操作前库存
	AfterStock  int        `json:"after_stock" gorm:"after_stock"`        // 操作后库存
	CreatedAt   *time.Time `json:"created_at" gorm:"created_at"`          // 创建时间
}

// TableName 表名称
func (*StockOperationItem) TableName() string {
	return "stock_operation_item"
}

// 库存日志表（保留兼容性，后续可考虑迁移）
type StockLog struct {
	ID           int64      `json:"id" gorm:"id,primaryKey;autoIncrement"` // 主键id
	ProductID    int64      `json:"product_id" gorm:"product_id"`          // 商品ID
	ProductName  string     `json:"product_name" gorm:"product_name"`      // 商品名称
	Types        int8       `json:"types" gorm:"types"`                    // 操作类型(1:入库,2:出库,3:退货)
	Quantity     int        `json:"quantity" gorm:"quantity"`              // 操作数量
	BeforeStock  int        `json:"before_stock" gorm:"before_stock"`      // 操作前库存
	AfterStock   int        `json:"after_stock" gorm:"after_stock"`        // 操作后库存
	OrderNo      string     `json:"order_no" gorm:"order_no"`              // 关联订单号(出库/退货时)
	Remark       string     `json:"remark" gorm:"remark"`                  // 备注
	Operator     string     `json:"operator" gorm:"operator"`              // 操作人
	OperatorID   int64      `json:"operator_id" gorm:"operator_id"`        // 操作人ID
	OperatorType int8       `json:"operator_type" gorm:"operator_type"`    // 操作人类型(1:用户,2:系统,3:管理员)
	BuyerName    string     `json:"buyer_name" gorm:"buyer_name"`          // 购买者名称(出库时)
	BuyerAccount string     `json:"buyer_account" gorm:"buyer_account"`    // 购买者账号(出库时)
	PurchaseTime *time.Time `json:"purchase_time" gorm:"purchase_time"`    // 购买时间(出库时)
	CreatedAt    *time.Time `json:"created_at" gorm:"created_at"`          // 创建时间
}

// TableName 表名称
func (*StockLog) TableName() string {
	return "stock_log"
}
