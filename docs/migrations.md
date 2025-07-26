# Database Migrations

This project uses [golang-migrate](https://github.com/golang-migrate/migrate) for database schema migrations.

## Migration Files

Migration files are stored in the `migrations/` directory with the following naming convention:

- `{version}_{name}.up.sql` - Forward migration
- `{version}_{name}.down.sql` - Rollback migration

Example:

```
migrations/
├── 000001_create_sellers_table.up.sql
├── 000001_create_sellers_table.down.sql
├── 000002_create_products_table.up.sql
└── 000002_create_products_table.down.sql
```

## Usage

### Environment Variables

Set the database connection URL:

```bash
export DATABASE_URL="postgres://username:password@localhost:5432/dbname?sslmode=disable"
```

### Running Migrations

**Apply all pending migrations:**

```bash
make migrate-up
```

**Rollback last migration:**

```bash
make migrate-down
```

**Rollback multiple migrations:**

```bash
make migrate-down STEPS=3
```

**Check current migration version:**

```bash
make migrate-version
```

**Create new migration files:**

```bash
make migrate-create NAME=add_user_preferences
```

### Manual Migration Tool

You can also use the migration CLI directly:

```bash
# Build the migration tool
make migrate-build

# Run migrations
./bin/migrate -database-url="$DATABASE_URL" -action=up

# Rollback migrations
./bin/migrate -database-url="$DATABASE_URL" -action=down -steps=1

# Check version
./bin/migrate -database-url="$DATABASE_URL" -action=version
```

## Automatic Migrations

The application automatically runs pending migrations on startup. This ensures your database schema is always up-to-date when the application starts.

To disable automatic migrations, modify the `cmd/marketplace/main.go` file and remove the migration code.

## Best Practices

1. **Always create both up and down migrations** - This allows for easy rollbacks
2. **Test your migrations** - Test both forward and backward migrations
3. **Keep migrations small** - One logical change per migration
4. **Never modify existing migrations** - Create new migrations to fix issues
5. **Use transactions when possible** - Wrap DDL statements in transactions where supported
6. **Backup before production migrations** - Always backup production data before running migrations

## Example Migration

**Create sellers table (up migration):**

```sql
-- migrations/000001_create_sellers_table.up.sql
CREATE TABLE IF NOT EXISTS sellers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_sellers_email ON sellers(email);
```

**Drop sellers table (down migration):**

```sql
-- migrations/000001_create_sellers_table.down.sql
DROP INDEX IF EXISTS idx_sellers_email;
DROP TABLE IF EXISTS sellers;
```
