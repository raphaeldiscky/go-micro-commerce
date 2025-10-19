CREATE TABLE user_addresses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    receiver_name VARCHAR(255) NOT NULL,
    address_line1 VARCHAR(255) NOT NULL,
    address_line2 VARCHAR(255),
    city VARCHAR(100) NOT NULL,
    state VARCHAR(100),
    postal_code VARCHAR(20) NOT NULL,
    country_code CHAR(2) NOT NULL, -- ISO 3166-1 alpha-2 code (e.g. 'JP', 'ID', 'US')
    latitude DECIMAL(10, 7),
    longitude DECIMAL(10, 7),
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    note TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    CONSTRAINT fk_user_addresses_user FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT chk_default_address CHECK (is_default IN (true, false)),
    CONSTRAINT chk_latitude_range CHECK (latitude >= -90 AND latitude <= 90),
    CONSTRAINT chk_longitude_range CHECK (longitude >= -180 AND longitude <= 180),
    CONSTRAINT chk_country_code_length CHECK (LENGTH(country_code) = 2),
    CONSTRAINT chk_postal_code_not_empty CHECK (postal_code <> '')
);

-- Ensure only one default address per user
CREATE UNIQUE INDEX ux_user_default_address
ON user_addresses (user_id)
WHERE is_default = TRUE;

-- Create trigger function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for user_addresses table
CREATE TRIGGER trigger_user_addresses_updated_at
    BEFORE UPDATE ON user_addresses
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();