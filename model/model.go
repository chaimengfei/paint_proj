package model

import (
	"encoding/json"
	"fmt"
	"math"
	"time"
)

// Product 商品表
type Product struct {
	ID            int64  `json:"id" gorm:"id,primaryKey;autoIncrement" ` // 主键ID
	Name          string `json:"name" gorm:"name"`                       // 商品全名
	SellerPrice   Amount `json:"seller_price" gorm:"seller_price"`       // 单价
	Cost          Amount `json:"cost" gorm:"cost"`                       // 成本价=运费成本+货物成本
	ShippingCost  Amount `json:"shipping_cost" gorm:"shipping_cost"`     // 运费成本
	ProductCost   Amount `json:"product_cost" gorm:"product_cost"`       // 货物成本
	CategoryId    int64  `json:"category_id" gorm:"category_id"`         // 分类ID
	Stock         int    `json:"stock" gorm:"stock"`                     // 库存
	Image         string `json:"image" gorm:"image"`                     // 图片地址
	Specification string `json:"specification" gorm:"specification"`     // 规格
	Unit          string `json:"unit" gorm:"unit"`                       // 单位
	Remark        string `json:"remark" gorm:"remark"`                   // 备注
	IsOnShelf     int8   `json:"is_on_shelf" gorm:"is_on_shelf"`         // 是否上架(1:上架,0:下架)
	ShopID        int64  `json:"shop_id" gorm:"shop_id"`                 // 关联店铺ID
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
	ShopID    int64  `json:"shop_id" gorm:"shop_id"`                 // 关联店铺ID
}

// TableName 表名称
func (*Category) TableName() string {
	return "category"
}

type Cart struct {
	ID        int64      `gorm:"id,primaryKey;autoIncrement" json:"id" `
	UserID    int64      `gorm:"column:user_id" json:"user_id"`
	ShopID    int64      `gorm:"column:shop_id" json:"shop_id"`
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
	OperatorTypeUser  = 1 // 用户
	OperatorTypeAdmin = 2 // 管理员

	// 用户来源类型
	UserSourceWechat = 1 // 小程序注册
	UserSourceAdmin  = 2 // 后台添加
	UserSourceMixed  = 3 // 混合（先后台添加，后小程序绑定）

	// 用户状态
	UserStatusDisabled = 0 // 禁用
	UserStatusEnabled  = 1 // 启用

	// 微信绑定状态
	WechatBindNo  = 0 // 未绑定微信
	WechatBindYes = 1 // 已绑定微信
)

// Order 订单表
type Order struct {
	ID              int64             `json:"id" gorm:"id,primaryKey;autoIncrement"`    // 主键id
	OrderNo         string            `json:"order_no" gorm:"order_no"`                 // 订单编号
	UserId          int64             `json:"user_id" gorm:"user_id"`                   // 用户ID
	ShopID          int64             `json:"shop_id" gorm:"shop_id"`                   // 关联店铺ID
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

	Items []StockOperationItem `json:"items" gorm:"-"` // ✅ 不映射到数据库，纯业务使用，现在使用StockOperationItem
}

// TableName 表名称
func (*Order) TableName() string {
	return "order"
}

// OrderLog 订单操作日志表
type OrderLog struct {
	ID           int64      `json:"id" gorm:"id,primaryKey;autoIncrement"` // 主键id
	OrderId      int64      `json:"order_id" gorm:"order_id"`              // 订单ID
	OrderNo      string     `json:"order_no" gorm:"order_no"`              // 订单编号
	Action       string     `json:"action" gorm:"action"`                  // 操作行为
	Operator     string     `json:"operator" gorm:"operator"`              // 操作人
	OperatorID   int64      `json:"operator_id" gorm:"operator_id"`        // 操作人ID
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

// User 用户表（支持小程序和后台管理系统）
type User struct {
	ID                int64     `json:"id" gorm:"id"`                                   // 用户ID
	Openid            string    `json:"openid" gorm:"openid"`                           // 微信OpenID
	Nickname          string    `json:"nickname" gorm:"nickname"`                       // 微信昵称（原始）
	Avatar            string    `json:"avatar" gorm:"avatar"`                           // 头像
	MobilePhone       string    `json:"mobile_phone" gorm:"mobile_phone"`               // 手机号（唯一标识）
	Source            int8      `json:"source" gorm:"source"`                           // 用户来源(1:小程序,2:后台添加,3:混合)
	IsEnable          int8      `json:"is_enable" gorm:"is_enable"`                     // 是否启用(1:启用,0:禁用) -防止恶意用户继续使用系统 - 处理用户投诉和纠纷时临时禁用 - 批量禁用测试账户或无效账户
	AdminDisplayName  string    `json:"admin_display_name" gorm:"admin_display_name"`   // 后台管理系统显示的客户名称
	WechatDisplayName string    `json:"wechat_display_name" gorm:"wechat_display_name"` // 微信小程序显示的客户名称
	HasWechatBind     int8      `json:"has_wechat_bind" gorm:"has_wechat_bind"`         // 是否已绑定微信(1:是,0:否)
	Remark            string    `json:"remark" gorm:"remark"`                           // 备注
	ShopID            int64     `json:"shop_id" gorm:"shop_id"`                         // 关联店铺ID
	CreatedAt         time.Time `json:"created_at" gorm:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" gorm:"updated_at"`
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
	ShopID         int64  `json:"shop_id" gorm:"shop_id"` // 关联店铺ID
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
	ID           int64  `json:"id" gorm:"id,primaryKey;autoIncrement"` // 主键id
	OperationNo  string `json:"operation_no" gorm:"operation_no"`      // 操作单号
	Types        int8   `json:"types" gorm:"types"`                    // 操作类型(1:入库,2:出库,3:退货)
	OutboundType int8   `json:"outbound_type" gorm:"outbound_type"`    // 出库类型(1:小程序购买,2:admin后台操作)
	Operator     string `json:"operator" gorm:"operator"`              // 操作人
	OperatorID   int64  `json:"operator_id" gorm:"operator_id"`        // 操作人ID
	OperatorType int8   `json:"operator_type" gorm:"operator_type"`    // 操作人类型(1:用户,2:系统,3:管理员)
	ShopID       int64  `json:"shop_id" gorm:"shop_id"`                // 关联店铺ID
	UserName     string `json:"user_name" gorm:"user_name"`            // 用户名称(出库时)
	UserID       int64  `json:"user_id" gorm:"user_id"`                // 用户ID(出库时)
	//UserAccount  string `json:"user_account" gorm:"user_account"`      // 用户账号(出库时)
	Remark              string            `json:"remark" gorm:"remark"`                               // 备注
	TotalAmount         Amount            `json:"total_amount" gorm:"total_amount"`                   // 总金额
	TotalQuantity       int               `json:"total_quantity" gorm:"total_quantity"`               // 总数量
	TotalProfit         Amount            `json:"total_profit" gorm:"total_profit"`                   // 总利润
	PaymentFinishStatus PaymentStatusCode `json:"payment_finish_status" gorm:"payment_finish_status"` // 支付完成状态(1:未支付,3:已支付)
	PaymentFinishTime   *time.Time        `json:"payment_finish_time" gorm:"payment_finish_time"`     // 支付完成时间
	Supplier            string            `json:"supplier" gorm:"supplier"`                           // 供货商
	CreatedAt           *time.Time        `json:"created_at" gorm:"created_at"`                       // 创建时间

	Items []StockOperationItem `json:"items" gorm:"-"` // 关联的子表数据（不映射到数据库）
}

// Supplier 供货商表
type Supplier struct {
	ID   int64  `json:"id" gorm:"primaryKey;autoIncrement"` // 供货商ID
	Name string `json:"name" gorm:"name;not null"`          // 供货商名称
	Area string `json:"area" gorm:"area"`                   // 供货商所在地区
}

// TableName 表名称
func (*StockOperation) TableName() string {
	return "stock_operation"
}

// TableName 表名称
func (*Supplier) TableName() string {
	return "supplier"
}

// 库存操作子表
type StockOperationItem struct {
	ID          int64  `json:"id" gorm:"id,primaryKey;autoIncrement"` // 主键id
	OperationID int64  `json:"operation_id" gorm:"operation_id"`      // 操作主表ID
	ShopID      int64  `json:"shop_id" gorm:"shop_id"`                // 关联店铺ID
	OrderID     int64  `json:"order_id" gorm:"order_id"`              // 关联订单ID(小程序购买时)
	OrderNo     string `json:"order_no" gorm:"order_no"`              // 关联订单号(小程序购买时)
	ProductID   int64  `json:"product_id" gorm:"product_id"`          // 商品ID

	Quantity      int        `json:"quantity" gorm:"quantity"`           // 操作数量
	UnitPrice     Amount     `json:"unit_price" gorm:"unit_price"`       // 单价
	TotalPrice    Amount     `json:"total_price" gorm:"total_price"`     // 总价
	BeforeStock   int        `json:"before_stock" gorm:"before_stock"`   // 操作前库存
	AfterStock    int        `json:"after_stock" gorm:"after_stock"`     // 操作后库存
	ProductCost   Amount     `json:"product_cost" gorm:"product_cost"`   // 货物成本(进价) 单位:分
	Profit        Amount     `json:"profit" gorm:"profit"`               // 利润(卖价-总成本)*数量 单位:分
	Remark        string     `json:"remark" gorm:"remark"`               // 备注
	ProductName   string     `json:"product_name" gorm:"product_name"`   // 商品全名
	Specification string     `json:"specification" gorm:"specification"` // 规格
	Unit          string     `json:"unit" gorm:"unit"`                   // 单位 L/桶/套
	CreatedAt     *time.Time `json:"created_at" gorm:"created_at"`       // 创建时间

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

// Shop 店铺表
type Shop struct {
	ID          int64     `json:"id" gorm:"id,primaryKey;autoIncrement"` // 店铺ID
	Name        string    `json:"name" gorm:"name"`                      // 店铺名称
	Code        string    `json:"code" gorm:"code"`                      // 店铺编码
	Address     string    `json:"address" gorm:"address"`                // 店铺地址
	Latitude    float64   `json:"latitude" gorm:"latitude"`              // 纬度
	Longitude   float64   `json:"longitude" gorm:"longitude"`            // 经度
	Phone       string    `json:"phone" gorm:"phone"`                    // 联系电话
	ManagerName string    `json:"manager_name" gorm:"manager_name"`      // 店长姓名
	IsActive    int8      `json:"is_active" gorm:"is_active"`            // 是否启用(1:启用,0:禁用)
	CreatedAt   time.Time `json:"created_at" gorm:"created_at"`          // 创建时间
	UpdatedAt   time.Time `json:"updated_at" gorm:"updated_at"`          // 更新时间
}

// TableName 表名称
func (*Shop) TableName() string {
	return "shop"
}

// 店铺常量
const (
	ShopYanjiao = 1 // 燕郊店
	ShopLaishui = 2 // 涞水店
)

// 地理位置相关请求结构
type LocationRequest struct {
	Latitude  float64 `json:"latitude" binding:"required"`  // 纬度
	Longitude float64 `json:"longitude" binding:"required"` // 经度
}

// 简化的店铺信息结构（用于列表展示）
type ShopSimple struct {
	ID          int64  `json:"id"`           // 店铺ID
	Name        string `json:"name"`         // 店铺名称
	Code        string `json:"code"`         // 店铺编码
	Address     string `json:"address"`      // 店铺地址
	Phone       string `json:"phone"`        // 联系电话
	ManagerName string `json:"manager_name"` // 店长姓名
	IsActive    int8   `json:"is_active"`    // 是否启用(1:启用,0:禁用)
}

// 后台管理员模型
type Operator struct {
	ID       int64  `json:"id" gorm:"id,primaryKey;autoIncrement"` // 管理员ID
	Name     string `json:"name" gorm:"name,uniqueIndex"`          // 管理员账号
	Password string `json:"-" gorm:"password"`                     // 密码(加密，不返回给前端)
	ShopID   int64  `json:"shop_id" gorm:"shop_id"`                // 所属店铺ID
	RealName string `json:"real_name" gorm:"real_name"`            // 真实姓名
	Phone    string `json:"phone" gorm:"phone"`                    // 联系电话
	IsActive int8   `json:"is_active" gorm:"is_active"`            // 是否启用(1:启用,0:禁用)
}

func (*Operator) TableName() string {
	return "operator"
}

// 后台管理员登录请求
type AdminLoginRequest struct {
	OperatorName string `json:"operator_name" binding:"required"` // 管理员账号
	Password     string `json:"password" binding:"required"`      // 密码
}

// 后台管理员登录响应
type AdminLoginResponse struct {
	Token     string       `json:"token"`      // JWT Token
	Operator  *Operator    `json:"operator"`   // 管理员信息
	ShopInfo  *ShopSimple  `json:"shop_info"`  // 店铺信息（普通管理员）
	ShopList  []ShopSimple `json:"shop_list"`  // 店铺列表（超级管理员）
	ExpiresIn int64        `json:"expires_in"` // Token 过期时间（秒）
}
