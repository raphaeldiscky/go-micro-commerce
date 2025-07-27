package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raphaeldiscky/go-ddd-template/internal/domain/entities"
	"github.com/raphaeldiscky/go-ddd-template/internal/domain/repositories"
	"github.com/raphaeldiscky/go-ddd-template/internal/infrastructure/db/sqlc"
)

type SellerRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

func NewSqlcSellerRepository(pool *pgxpool.Pool) repositories.SellerRepository {
	return &SellerRepository{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}

func (repo *SellerRepository) Create(seller *entities.ValidatedSeller) (*entities.Seller, error) {
	ctx := context.Background()

	now := time.Now()
	params := sqlc.CreateSellerParams{
		ID:        seller.Id,
		Name:      seller.Name,
		Email:     seller.Email,
		CreatedAt: now,
		UpdatedAt: now,
	}

	dbSeller, err := repo.queries.CreateSeller(ctx, params)
	if err != nil {
		return nil, err
	}

	// Convert sqlc model to domain entity
	return fromSqlcSeller(&dbSeller), nil
}

func (repo *SellerRepository) FindById(id uuid.UUID) (*entities.Seller, error) {
	ctx := context.Background()

	dbSeller, err := repo.queries.GetSellerByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return fromSqlcSeller(&dbSeller), nil
}

func (repo *SellerRepository) FindAll() ([]*entities.Seller, error) {
	ctx := context.Background()

	dbSellers, err := repo.queries.ListSellers(ctx)
	if err != nil {
		return nil, err
	}

	sellers := make([]*entities.Seller, len(dbSellers))
	for i, dbSeller := range dbSellers {
		sellers[i] = fromSqlcSeller(&dbSeller)
	}

	return sellers, nil
}

func (repo *SellerRepository) Update(seller *entities.ValidatedSeller) (*entities.Seller, error) {
	ctx := context.Background()

	params := sqlc.UpdateSellerParams{
		ID:        seller.Id,
		Name:      seller.Name,
		Email:     seller.Email,
		UpdatedAt: time.Now(),
	}

	dbSeller, err := repo.queries.UpdateSeller(ctx, params)
	if err != nil {
		return nil, err
	}

	return fromSqlcSeller(&dbSeller), nil
}

func (repo *SellerRepository) Delete(id uuid.UUID) error {
	ctx := context.Background()
	return repo.queries.DeleteSeller(ctx, id)
}

// Helper function to convert sqlc model to domain entity
func fromSqlcSeller(dbSeller *sqlc.Seller) *entities.Seller {
	return &entities.Seller{
		Id:        dbSeller.ID,
		CreatedAt: dbSeller.CreatedAt,
		UpdatedAt: dbSeller.UpdatedAt,
		Name:      dbSeller.Name,
		Email:     dbSeller.Email,
	}
}
