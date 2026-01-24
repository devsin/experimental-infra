package transfer

import (
	"time"

	"github.com/google/uuid"
)

// Transfer is the domain model for a money movement between accounts.
type Transfer struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	FromAccountID uuid.UUID `gorm:"type:uuid;index" json:"from_account_id"`
	ToAccountID   uuid.UUID `gorm:"type:uuid;index" json:"to_account_id"`
	AmountCents   int64     `json:"amount_cents"`
	CreatedAt     time.Time `json:"created_at"`
}
