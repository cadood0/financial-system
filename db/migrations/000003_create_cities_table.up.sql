CREATE TABLE cities (
    id         BIGSERIAL PRIMARY KEY,
    name       CITEXT NOT NULL,
    region     VARCHAR(100) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX cities_name_unique ON cities (name) WHERE deleted_at IS NULL;