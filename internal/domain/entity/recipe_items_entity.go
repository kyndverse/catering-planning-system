package entity

type RecipeItem struct {
	ID           int64
	RecipeID     string
	IngredientID *string
	SubRecipeID  *float64
	Quantity     string

	Recipe     Recipe
	Ingredient *Ingredient
	SubRecipe  *Recipe
}
