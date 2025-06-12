package model

import "time"

// Product 油漆表
type Product struct {
	ID           int64   `json:"id" gorm:"id"`                       // 主键id
	Name         string  `json:"name" gorm:"name"`                   // 商品名
	SellerPrice  float64 `json:"seller_price" gorm:"seller_price"`   // 卖价
	Cost         float64 `json:"cost" gorm:"cost"`                   // 成本价(暂不用)
	ShippingCost float64 `json:"shipping_cost" gorm:"shipping_cost"` // 运费(暂不用)
	ProductCost  float64 `json:"product_cost" gorm:"product_cost"`   // 货物成本(暂不用)
	CategoryId   int64   `json:"category_id" gorm:"category_id"`     // 分类id
	Image        string  `json:"image" gorm:"image"`                 // 图片地址
	Unit         string  `json:"unit" gorm:"unit"`                   // 单位 L/桶/套
	Remark       string  `json:"remark" gorm:"remark"`               // 备注
}

// TableName 表名称
func (*Product) TableName() string {
	return "product"
}

// Category 商品分类表
type Category struct {
	ID        int64  `json:"id" gorm:"id"`                 // 分类ID
	Name      string `json:"name" gorm:"name"`             // 分类名称
	SortOrder int64  `json:"sort_order" gorm:"sort_order"` // 排序权重(数字越大越靠前)
}

// TableName 表名称
func (*Category) TableName() string {
	return "category"
}

type Cart struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	UserID    int64     `gorm:"column:user_id" json:"user_id"`
	ProductID int64     `gorm:"column:product_id" json:"product_id"`
	Quantity  int       `gorm:"column:quantity" json:"quantity"`
	Selected  bool      `gorm:"column:selected" json:"selected"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (Cart) TableName() string {
	return "cart"
}
