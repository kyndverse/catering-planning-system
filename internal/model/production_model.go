package model

import "time"

// ProductionRequirementRow is the unformatted result of the production query.
// Numeric values intentionally remain strings so no precision is lost at the DB boundary.
type ProductionRequirementRow struct {
	IngredientID           int64
	Name                   string
	BaseUnit               string
	PurchaseUnit           string
	BaseQtyPerPurchaseUnit string
	RequiredQty            string
	StockQty               string
	PricePerPurchaseUnit   *string
}

type ProductionMenu struct {
	MenuID   int64
	MenuName string
	Portions int
}

type ProductionDateSummary struct {
	Date          time.Time
	OrderCount    int
	TotalPortions int
	MaxEventDate  time.Time
	AllLocked     bool
}

type ProductionRequirement struct {
	Name     string
	Unit     string
	Required string
	Stock    string
	Short    bool
}

type ShoppingItem struct {
	Name                 string
	Unit                 string
	Shortage             string
	BuyQty               int64
	PurchaseUnit         string
	HasPrice             bool
	PricePerPurchaseUnit string
	Subtotal             string
}

type ShoppingList struct {
	Items           []ShoppingItem
	TotalCost       string
	PriceIncomplete bool
}

type ProductionPlan struct {
	Date              time.Time
	Orders            []OrderSummary
	Menus             []ProductionMenu
	TotalPortions     int
	HasEditableOrders bool
	Requirements      []ProductionRequirement
	Shopping          ShoppingList
}
