package dto

type AddToCartRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"gte=1"`
}

type CartProductResponse struct {
	ID       uint    `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	ImageURL string  `json:"image_url"` // ✅ add this
}

type CartItemResponse struct {
	ID       uint                 `json:"id"` // ✅ cart item id
	Quantity int                  `json:"quantity"`
	Product  CartProductResponse `json:"product"`
}