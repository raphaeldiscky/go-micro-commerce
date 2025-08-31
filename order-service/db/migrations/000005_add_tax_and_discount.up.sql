-- Add tax_total and discount_total columns to orders table
ALTER TABLE orders 
ADD COLUMN IF NOT EXISTS tax_total DECIMAL(10, 2) NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS discount_total DECIMAL(10, 2) NOT NULL DEFAULT 0;

-- Add tax and discount to order_items
ALTER TABLE order_items 
ADD COLUMN IF NOT EXISTS tax DECIMAL(10, 2) NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS discount DECIMAL(10, 2) NOT NULL DEFAULT 0;