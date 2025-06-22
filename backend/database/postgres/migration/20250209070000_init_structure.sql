-- +goose Up
-- +goose StatementBegin
CREATE TABLE users
(
    id           BIGSERIAL PRIMARY KEY,
    oid          VARCHAR(64) NOT NULL UNIQUE,
    firstname    VARCHAR(255) NOT NULL,
    lastname     VARCHAR(255) NOT NULL,
    email        VARCHAR(255) UNIQUE,
    photo_url    TEXT,
    created_at   TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- * auto-update function for updated_at timestamps
CREATE OR REPLACE FUNCTION auto_updated_at()
    RETURNS TRIGGER AS
$$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- * triggers to automatically update updated_at
CREATE TRIGGER auto_updated_at_users
    BEFORE UPDATE
    ON users
    FOR EACH ROW
EXECUTE FUNCTION auto_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users;
DROP FUNCTION auto_updated_at;
-- +goose StatementEnd
