CREATE TABLE stock_movements (
  id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  ingredient_id INT NOT NULL REFERENCES ingredients (id),
  movement_date DATE NOT NULL,
  type TEXT NOT NULL CHECK (type IN ('PURCHASE', 'USAGE', 'ADJUSTMENT')),
  quantity NUMERIC NOT NULL, -- PURCHASE (+), USAGE (-), ADJUSTMENT (+/-)
  total_cost NUMERIC, -- filled for PURCHASE
  production_date DATE, -- filled for USAGE
  note TEXT,
  CHECK (
    type <> 'PURCHASE'
    OR quantity > 0
  ),
  CHECK (
    type <> 'USAGE'
    OR (
      quantity < 0
      AND production_date IS NOT NULL
    )
  )
);

CREATE INDEX idx_stock_movements_ingredient_id ON stock_movements (ingredient_id);

CREATE INDEX idx_stock_movements_usage_date ON stock_movements (production_date)
WHERE
  type = 'USAGE';