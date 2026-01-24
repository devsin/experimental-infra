CREATE TABLE IF NOT EXISTS transfers (
    id UUID PRIMARY KEY,
    from_account_id UUID NOT NULL,
    to_account_id UUID NOT NULL,
    amount_cents BIGINT NOT NULL CHECK (amount_cents > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_transfers_from_account ON transfers (from_account_id);
CREATE INDEX IF NOT EXISTS idx_transfers_to_account ON transfers (to_account_id);
