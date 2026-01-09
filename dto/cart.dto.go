package dto

type AddToCartRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"gte=1"`
}

type CartProductResponse struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

type CartItemResponse struct {
	Quantity int                 `json:"quantity"`
	Product  CartProductResponse `json:"product"`
}
