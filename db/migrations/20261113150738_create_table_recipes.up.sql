CREATE TABLE recipes (
  id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  type TEXT NOT NULL CHECK (type IN ('PREPARED', 'COMPONENT', 'MENU')),
  yield_qty NUMERIC NOT NULL CHECK (yield_qty > 0), -- how much this recipe produces...
  yield_unit TEXT NOT NULL -- ...'portion' / 'g' / 'ml'
);

CREATE TABLE recipe_items (
  id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  recipe_id INT NOT NULL REFERENCES recipes (id) ON DELETE CASCADE,
  ingredient_id INT REFERENCES ingredients (id),
  sub_recipe_id INT REFERENCES recipes (id),
  quantity NUMERIC NOT NULL CHECK (quantity > 0),
  CHECK (
    (ingredient_id IS NULL) <> (sub_recipe_id IS NULL)
  ), -- exactly one
  CHECK (
    sub_recipe_id IS DISTINCT
    FROM
      recipe_id
  ) -- no direct self-reference
);

CREATE INDEX idx_recipe_items_recipe_id ON recipe_items (recipe_id);