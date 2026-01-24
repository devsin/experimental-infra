package account

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository defines data access for accounts.
type Repository interface {
	Create(context.Context, *Account) error
	Get(context.Context, uuid.UUID) (*Account, error)
	Update(context.Context, uuid.UUID, map[string]any) (*Account, error)
	Delete(context.Context, uuid.UUID) error
	Migrate(context.Context) error
}

type gormRepository struct {
	db *gorm.DB
}

// NewRepository returns a GORM-backed repository.
func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) Create(ctx context.Context, acct *Account) error {
	return r.db.WithContext(ctx).Create(acct).Error
}

func (r *gormRepository) Get(ctx context.Context, id uuid.UUID) (*Account, error) {
	var acct Account
	if err := r.db.WithContext(ctx).First(&acct, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &acct, nil
}

func (r *gormRepository) Update(ctx context.Context, id uuid.UUID, updates map[string]any) (*Account, error) {
	if err := r.db.WithContext(ctx).Model(&Account{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, err
	}
	return r.Get(ctx, id)
}

func (r *gormRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&Account{}, "id = ?", id).Error
}

func (r *gormRepository) Migrate(ctx context.Context) error {
	return r.db.WithContext(ctx).AutoMigrate(&Account{})
}
