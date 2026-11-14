package model

type IngredientResponse struct {
	ID                     int64
	Name                   string
	BaseUnit               string
	PurchaseUnit           string
	BaseQtyPerPurchaseUnit int64
}

type CreateIngredientRequest struct {
	ID                     int64
	Name                   string `form:"name" validate:"required"`
	BaseUnit               string `form:"base_unit" validate:"required"`
	PurchaseUnit           string `form:"purchase_unit" validate:"required"`
	BaseQtyPerPurchaseUnit int64  `form:"base_qty_per_purchase_unit" validate:"required"`
}
