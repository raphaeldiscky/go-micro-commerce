package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raphaeldiscky/go-ddd-template/internal/domain/entities"
	"github.com/raphaeldiscky/go-ddd-template/internal/domain/repositories"
	"github.com/raphaeldiscky/go-ddd-template/internal/infrastructure/db/sqlc"
)

type SqlcProductRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

func NewSqlcProductRepository(pool *pgxpool.Pool) repositories.ProductRepository {
	return &SqlcProductRepository{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}

func (repo *SqlcProductRepository) Create(product *entities.ValidatedProduct) (*entities.Product, error) {
	ctx := context.Background()

	now := time.Now()

	// Convert price to pgtype.Numeric
	priceStr := fmt.Sprintf("%.2f", product.Price)
	priceNumeric := pgtype.Numeric{}
	if err := priceNumeric.Scan(priceStr); err != nil {
		return nil, fmt.Errorf("failed to convert price: %w", err)
	}

	params := sqlc.CreateProductParams{
		ID:          product.Id,
		Name:        product.Name,
		Description: pgtype.Text{String: "", Valid: false}, // Use empty description for now
		Price:       priceNumeric,
		SellerID:    uuid.UUID{}, // Need to set this properly
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	dbProduct, err := repo.queries.CreateProduct(ctx, params)
	if err != nil {
		return nil, err
	}

	// Convert sqlc model to domain entity
	return fromSqlcProduct(&dbProduct)
}

func (repo *SqlcProductRepository) FindById(id uuid.UUID) (*entities.Product, error) {
	ctx := context.Background()

	dbProduct, err := repo.queries.GetProductByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return fromSqlcProduct(&dbProduct)
}

func (repo *SqlcProductRepository) FindAll() ([]*entities.Product, error) {
	ctx := context.Background()

	dbProducts, err := repo.queries.ListProducts(ctx)
	if err != nil {
		return nil, err
	}

	products := make([]*entities.Product, len(dbProducts))
	for i, dbProduct := range dbProducts {
		product, err := fromSqlcProduct(&dbProduct)
		if err != nil {
			return nil, err
		}
		products[i] = product
	}

	return products, nil
}

func (repo *SqlcProductRepository) Update(product *entities.ValidatedProduct) (*entities.Product, error) {
	ctx := context.Background()

	// Convert price to pgtype.Numeric
	priceStr := fmt.Sprintf("%.2f", product.Price)
	priceNumeric := pgtype.Numeric{}
	if err := priceNumeric.Scan(priceStr); err != nil {
		return nil, fmt.Errorf("failed to convert price: %w", err)
	}

	params := sqlc.UpdateProductParams{
		ID:          product.Id,
		Name:        product.Name,
		Description: pgtype.Text{String: "", Valid: false},
		Price:       priceNumeric,
		SellerID:    uuid.UUID{}, // Need to set this properly
		UpdatedAt:   time.Now(),
	}

	dbProduct, err := repo.queries.UpdateProduct(ctx, params)
	if err != nil {
		return nil, err
	}

	return fromSqlcProduct(&dbProduct)
}

func (repo *SqlcProductRepository) Delete(id uuid.UUID) error {
	ctx := context.Background()
	return repo.queries.DeleteProduct(ctx, id)
}

// Helper function to convert sqlc model to domain entity
func fromSqlcProduct(dbProduct *sqlc.Product) (*entities.Product, error) {
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
	seller := entities.Seller{
		Id:        dbProduct.SellerID,
		CreatedAt: time.Now(), // This should be fetched from the seller table
		UpdatedAt: time.Now(), // This should be fetched from the seller table
		Name:      "",         // This should be fetched from the seller table
		Email:     "",         // This should be fetched from the seller table
	}

	return &entities.Product{
		Id:        dbProduct.ID,
		CreatedAt: dbProduct.CreatedAt,
		UpdatedAt: dbProduct.UpdatedAt,
		Name:      dbProduct.Name,
		Price:     price,
		Seller:    seller,
	}, nil
}
