CREATE TABLE ingredients (
  id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  base_unit TEXT NOT NULL CHECK (base_unit IN ('g', 'ml', 'pcs')),
  purchase_unit TEXT NOT NULL, -- 'sack 5 kg', 'pouch 2 L', 'kg'
  base_qty_per_purchase_unit NUMERIC NOT NULL CHECK (base_qty_per_purchase_unit > 0) -- 5000 (g), 2000 (ml), 1000 (g)
);