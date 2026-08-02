CREATE TABLE monthly_charges (
    id           BIGSERIAL PRIMARY KEY,
    member_id    BIGINT NOT NULL REFERENCES members(id) ON DELETE RESTRICT,
    fee_type_id  BIGINT NOT NULL REFERENCES fee_types(id) ON DELETE RESTRICT,
    amount_cents BIGINT NOT NULL CHECK (amount_cents > 0),
    start_date   DATE NOT NULL,
    end_date     DATE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ,

    CONSTRAINT charges_dates_valid CHECK (end_date IS NULL OR end_date > start_date)
);

CREATE UNIQUE INDEX charges_member_feetype_unique
    ON monthly_charges (member_id, fee_type_id)
    WHERE deleted_at IS NULL;

CREATE INDEX charges_member_id_idx ON monthly_charges (member_id);
CREATE INDEX charges_fee_type_id_idx ON monthly_charges (fee_type_id);