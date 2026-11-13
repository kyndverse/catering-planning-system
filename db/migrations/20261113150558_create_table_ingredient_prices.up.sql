CREATE TABLE ingredient_prices (
  id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  ingredient_id INT NOT NULL REFERENCES ingredients (id),
  price_per_purchase_unit NUMERIC NOT NULL CHECK (price_per_purchase_unit >= 0),
  effective_from DATE NOT NULL,
  UNIQUE (ingredient_id, effective_from)
);