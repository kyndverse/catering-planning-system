package model

type RecipeResponse struct {
	ID        int64
	Name      string
	Type      string
	YieldQty  float64
	YieldUnit string
}

type RecipeSummary struct {
	ID        int64
	Name      string
	Type      string
	YieldQty  float64
	YieldUnit string
	ItemCount int
}

type CreateRecipeRequest struct {
	Name      string  `form:"name" validate:"required"`
	Type      string  `form:"type" validate:"required"`
	YieldQty  float64 `form:"yield_qty" validate:"required"`
	YieldUnit string  `form:"yield_unit" validate:"required"`
}

type UpdateRecipeRequest struct {
	ID        int64
	Name      string  `form:"name"`
	Type      string  `form:"type"`
	YieldQty  float64 `form:"yield_qty"`
	YieldUnit string  `form:"yield_unit"`
}

type DeleteRecipeRequest struct {
	ID int64
}

type SearchRecipeRequest struct {
	Type string
}

type RecipeItemView struct {
	ID       int64
	Kind     string // "ingredient" | "recipe"
	TargetID int64
	Name     string
	Quantity string
	Unit     string
}

type AddRecipeItemRequest struct {
	RecipeID int64
	Item     string `form:"item" validate:"required"`     // "ingredient:4" atau "recipe:2"
	Quantity string `form:"quantity" validate:"required"` // string desimal, tanpa float
}

type SubRecipeItemOption struct {
	ID   int64
	Name string
	Unit string
}

type IngredientRequirement struct {
	IngredientName string
	Quantity       string
	Unit           string
}

const (
	ItemKindIngredient = "ingredient"
	ItemKindRecipe     = "recipe"
)

// UpdateRecipeItemRequest: ubah jumlah satu isi resep.
// RecipeID dan ItemID berasal dari path, BUKAN dari form, maka form:"-".
type UpdateRecipeItemRequest struct {
	RecipeID int64  `form:"-"`
	ItemID   int64  `form:"-"`
	Quantity string `form:"quantity" validate:"required"`
}

// DeleteRecipeItemRequest: hapus satu isi resep (keduanya dari path).
type DeleteRecipeItemRequest struct {
	RecipeID int64
	ItemID   int64
}