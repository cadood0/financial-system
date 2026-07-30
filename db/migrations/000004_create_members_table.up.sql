CREATE TABLE members (
    id         BIGSERIAL PRIMARY KEY,
    full_name  VARCHAR(150) NOT NULL,
    phone      VARCHAR(20) NOT NULL,
    city_id    BIGINT NOT NULL REFERENCES cities(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX members_phone_unique ON members (phone) WHERE deleted_at IS NULL;
CREATE INDEX members_city_id_idx ON members (city_id);