CREATE TABLE payments (
    id           BIGSERIAL PRIMARY KEY,
    member_id    BIGINT NOT NULL REFERENCES members(id) ON DELETE RESTRICT,
    charge_id    BIGINT NOT NULL REFERENCES monthly_charges(id) ON DELETE RESTRICT,
    recorded_by  BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    period       DATE NOT NULL CHECK (EXTRACT(DAY FROM period) = 1),
    amount_cents BIGINT NOT NULL CHECK (amount_cents > 0),
    note         TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX payments_charge_period_idx ON payments (charge_id, period);
CREATE INDEX payments_member_id_idx ON payments (member_id);