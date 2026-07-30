CREATE TABLE fee_types (
    id         BIGSERIAL PRIMARY KEY,
    name       CITEXT NOT NULL,
    description  TEXT NOT NULL DEFAULT '',
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX fee_types_name_unique
ON fee_types(name)
WHERE deleted_at IS NULL;