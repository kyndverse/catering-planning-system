package entity

type Ingredient struct {
	ID                     int64
	Name                   string
	BaseUnit               string
	PurchaseUnit           string
	BaseQtyPerPurchaseUnit int64
}
