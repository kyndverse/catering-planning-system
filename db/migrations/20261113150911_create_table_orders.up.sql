CREATE TABLE orders (
  id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  customer_name TEXT NOT NULL,
  event_date DATE NOT NULL,
  production_date DATE NOT NULL, -- default in app: = event_date
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CHECK (production_date <= event_date)
);

CREATE INDEX idx_orders_production_date ON orders (production_date);

CREATE TABLE order_items (
  order_id INT NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
  menu_id INT NOT NULL REFERENCES recipes (id), -- recipe of type MENU
  portions INT NOT NULL CHECK (portions > 0),
  PRIMARY KEY (order_id, menu_id)
);