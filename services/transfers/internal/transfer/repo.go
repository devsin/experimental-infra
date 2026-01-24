package transfer

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository defines data access for transfers.
type Repository interface {
	Create(context.Context, *Transfer) error
	Get(context.Context, uuid.UUID) (*Transfer, error)
	List(context.Context, *uuid.UUID, int) ([]Transfer, error)
	Migrate(context.Context) error
}

type gormRepository struct {
	db *gorm.DB
}

// NewRepository returns a GORM-backed repository.
func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) Create(ctx context.Context, t *Transfer) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *gormRepository) Get(ctx context.Context, id uuid.UUID) (*Transfer, error) {
	var t Transfer
	if err := r.db.WithContext(ctx).First(&t, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *gormRepository) List(ctx context.Context, accountID *uuid.UUID, limit int) ([]Transfer, error) {
	q := r.db.WithContext(ctx).Order("created_at DESC").Limit(limit)
	if accountID != nil {
		q = q.Where("from_account_id = ? OR to_account_id = ?", accountID, accountID)
	}

	var transfers []Transfer
	if err := q.Find(&transfers).Error; err != nil {
		return nil, err
	}
	return transfers, nil
}

func (r *gormRepository) Migrate(ctx context.Context) error {
	return r.db.WithContext(ctx).AutoMigrate(&Transfer{})
}
