package cart

type AddToCartRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"gte=1"`
}

type CartProductResponse struct {
	ID       uint    `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	ImageURL string  `json:"image_url"`
}

type CartItemResponse struct {
	ID       uint                `json:"id"`
	Quantity int                 `json:"quantity"`
	Product  CartProductResponse `json:"product"`
}
type CartResponse struct {
	CartItems []CartItemResponse `json:"cartItems"`
	Total     float64            `josn:"total"`
}
