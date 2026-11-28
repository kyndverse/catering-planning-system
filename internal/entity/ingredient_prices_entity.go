package entity

import "time"

type IngredientPrice struct {
	ID                   int64
	IngredientID         int64
	PricePerPurchaseUnit float64
	EffectiveFrom        time.Time
}
