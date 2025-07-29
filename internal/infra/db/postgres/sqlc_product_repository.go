package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	entity "github.com/raphaeldiscky/go-ddd-template/internal/domain/entity"
	repository "github.com/raphaeldiscky/go-ddd-template/internal/domain/repository"
	"github.com/raphaeldiscky/go-ddd-template/internal/infra/db/sqlc"
)

// SqlcProductRepository implements the repository.ProductRepository interface using SQLC.
type SqlcProductRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

// NewSqlcProductRepository creates a new instance of SqlcProductRepository.
func NewSqlcProductRepository(pool *pgxpool.Pool) repository.ProductRepository {
	return &SqlcProductRepository{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}

// Create adds a new product to the database.
func (repo *SqlcProductRepository) Create(
	product *entity.ValidatedProduct,
) (*entity.Product, error) {
	ctx := context.Background()

	now := time.Now()

	// Convert price to pgtype.Numeric
	priceStr := fmt.Sprintf("%.2f", product.Price)
	priceNumeric := pgtype.Numeric{}

	if err := priceNumeric.Scan(priceStr); err != nil {
		return nil, fmt.Errorf("failed to convert price: %w", err)
	}

	params := sqlc.CreateProductParams{
		ID:          product.ID,
		Name:        product.Name,
		Description: pgtype.Text{String: "", Valid: false}, // Use empty description for now
		Price:       priceNumeric,
		SellerID:    uuid.UUID{}, // Need to set this properly
		CreatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
	}

	dbProduct, err := repo.queries.CreateProduct(ctx, params)
	if err != nil {
		return nil, err
	}

	// Convert sqlc model to domain entity
	return fromSqlcProduct(&dbProduct)
}

// FindByID retrieves a product by ID.
func (repo *SqlcProductRepository) FindByID(id uuid.UUID) (*entity.Product, error) {
	ctx := context.Background()

	dbProduct, err := repo.queries.GetProductByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return fromSqlcProduct(&dbProduct)
}

// FindAll retrieves all products.
func (repo *SqlcProductRepository) FindAll() ([]*entity.Product, error) {
	ctx := context.Background()

	dbProducts, err := repo.queries.ListProducts(ctx)
	if err != nil {
		return nil, err
	}

	products := make([]*entity.Product, len(dbProducts))

	for i := range dbProducts {
		product, err := fromSqlcProduct(&dbProducts[i])
		if err != nil {
			return nil, err
		}

		products[i] = product
	}

	return products, nil
}

// Update modifies an existing product.
func (repo *SqlcProductRepository) Update(
	product *entity.ValidatedProduct,
) (*entity.Product, error) {
	ctx := context.Background()

	// Convert price to pgtype.Numeric
	priceStr := fmt.Sprintf("%.2f", product.Price)
	priceNumeric := pgtype.Numeric{}

	if err := priceNumeric.Scan(priceStr); err != nil {
		return nil, fmt.Errorf("failed to convert price: %w", err)
	}

	params := sqlc.UpdateProductParams{
		ID:          product.ID,
		Name:        product.Name,
		Description: pgtype.Text{String: "", Valid: false},
		Price:       priceNumeric,
		SellerID:    uuid.UUID{}, // Need to set this properly
		UpdatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}

	dbProduct, err := repo.queries.UpdateProduct(ctx, params)
	if err != nil {
		return nil, err
	}

	return fromSqlcProduct(&dbProduct)
}

// Delete removes a product by ID.
func (repo *SqlcProductRepository) Delete(id uuid.UUID) error {
	ctx := context.Background()

	return repo.queries.DeleteProduct(ctx, id)
}

// fromSqlcProduct converts a sqlc.Product to an entity.Product.
func fromSqlcProduct(dbProduct *sqlc.Product) (*entity.Product, error) {
	// Convert pgtype.Numeric to float64
	var price float64

	if dbProduct.Price.Valid {
		// Get the float64 value from pgtype.Numeric
		f64, err := dbProduct.Price.Float64Value()
		if err != nil {
			return nil, fmt.Errorf("failed to parse price: %w", err)
		}

		price = f64.Float64
	}

	// For now, we'll create a simple seller struct - this needs to be handled properly
	// by fetching the seller separately or using joins
	seller := entity.Seller{
		ID:        dbProduct.SellerID,
		CreatedAt: time.Now(), // This should be fetched from the seller table
		UpdatedAt: time.Now(), // This should be fetched from the seller table
		Name:      "",         // This should be fetched from the seller table
		Email:     "",         // This should be fetched from the seller table
	}

	return &entity.Product{
		ID:        dbProduct.ID,
		CreatedAt: dbProduct.CreatedAt.Time,
		UpdatedAt: dbProduct.UpdatedAt.Time,
		Name:      dbProduct.Name,
		Price:     price,
		Seller:    seller,
	}, nil
}
