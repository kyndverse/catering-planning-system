package repository

import (
	"database/sql"
	"golang-web/internal/domain/entity"

	"github.com/sirupsen/logrus"
)

type IngredientPriceRepository struct {
	Log *logrus.Logger
}

func NewIngredientPriceRepository(logger *logrus.Logger) *IngredientPriceRepository {
	return &IngredientPriceRepository{
		Log: logger,
	}
}

func (p *IngredientPriceRepository) FindAllByIngedientID(tx *sql.Tx, id int64) ([]entity.IngredientPrice, error) {
	query := `SELECT id, ingredient_id, price_per_purchase_unit, effective_from FROM ingredient_prices WHERE ingredient_id=$1 ORDER BY effective_from DESC`

	var prices []entity.IngredientPrice
	rows, err := tx.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		price := entity.IngredientPrice{}
		if err := rows.Scan(
			&price.ID,
			&price.IngredientID,
			&price.PricePerPurchaseUnit,
			&price.EffectiveFrom,
		); err != nil {
			return nil, err
		}

		prices = append(prices, price)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return prices, nil
}

func (p *IngredientPriceRepository) Create(tx *sql.Tx, price *entity.IngredientPrice) error {
	query := `
        INSERT INTO ingredient_prices (ingredient_id, price_per_purchase_unit, effective_from)
        VALUES ($1, $2, $3)
        RETURNING id
    `

	return tx.QueryRow(
		query,
		price.IngredientID,
		price.PricePerPurchaseUnit,
		price.EffectiveFrom,
	).Scan(&price.ID)
}
