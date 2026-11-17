package model

type RecipeResponse struct {
	ID        int64
	Name      string
	Type      string
	YieldQty  float64
	YieldUnit string
}

type RecipeSummary struct {
	ID         int64
	Name       string
	Type       string
	YieldQty   float64
	YieldUnit  string
	ItemCount int
}

type CreateRecipeRequest struct {
	Name      string  `form:"name" validate:"required"`
	Type      string  `form:"type" validate:"required"`
	YieldQty  float64 `form:"yield_qty" validate:"required"`
	YieldUnit string  `form:"yield_unit" validate:"required"`
}

type SearchRecipeRequest struct {
	Type string
}
