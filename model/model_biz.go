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
