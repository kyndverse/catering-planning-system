package model

import "time"

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

type UpdateIngredientRequest struct {
	ID                     int64
	Name                   string `form:"name"`
	BaseUnit               string `form:"base_unit"`
	PurchaseUnit           string `form:"purchase_unit"`
	BaseQtyPerPurchaseUnit int64  `form:"base_qty_per_purchase_unit"`
}

type CreateIngredientPriceRequest struct {
	IngredientID         int64
	PricePerPurchaseUnit float64   `form:"price_per_purchase_unit" validate:"required"`
	EffectiveFrom        time.Time `validate:"required"`
}

type IngredientPriceResponse struct {
	ID                   int64
	IngredientID         int64
	PricePerPurchaseUnit float64
	EffectiveFrom        time.Time
}

type DeleteIngredientRequest struct {
	ID int64
}
