package service

import (
	"golang-web/internal/model"
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

// Menolak "1e3", "1/3", "-5", "abc".
var itemQtyPattern = regexp.MustCompile(`^\d{1,12}(\.\d{1,6})?$`)

// parseItemRef mengurai "ingredient:4" / "recipe:2".
func parseItemRef(v string) (kind string, id int64, ok bool) {
	k, idStr, found := strings.Cut(strings.TrimSpace(v), ":")
	if !found || (k != model.ItemKindIngredient && k != model.ItemKindRecipe) {
		return "", 0, false
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return "", 0, false
	}
	return k, id, true
}

// parsePositiveQuantity memastikan string desimal valid dan > 0.
func parsePositiveQuantity(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	if !itemQtyPattern.MatchString(raw) {
		return "", false
	}
	r, ok := new(big.Rat).SetString(raw)
	if !ok || r.Sign() <= 0 {
		return "", false
	}
	return raw, true
}
