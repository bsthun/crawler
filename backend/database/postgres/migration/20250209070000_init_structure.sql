-- +goose Up
-- +goose StatementBegin
CREATE TABLE users
(
    id         BIGSERIAL PRIMARY KEY,
    oid        VARCHAR(64)  NOT NULL UNIQUE,
    firstname  VARCHAR(255) NOT NULL,
    lastname   VARCHAR(255) NOT NULL,
    email      VARCHAR(255) UNIQUE,
    photo_url  TEXT,
    created_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE categories
(
    id         BIGSERIAL PRIMARY KEY,
    name       VARCHAR(255) CHECK ( name ~ '^[a-z0-9_]+$' ) NOT NULL UNIQUE,
    created_at TIMESTAMP                                    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP                                    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE uploads
(
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT    REFERENCES users (id) ON DELETE SET NULL NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE tasks
(
    id            BIGSERIAL PRIMARY KEY,
    user_id       BIGINT                                                                           REFERENCES users (id) ON DELETE SET NULL NULL,
    upload_id     BIGINT                                                                           REFERENCES uploads (id) ON DELETE SET NULL NULL,
    category_id   BIGINT                                                                           REFERENCES categories (id) ON DELETE SET NULL NULL,
    type          VARCHAR(64) CHECK ( type IN ('web', 'doc', 'youtube') )                          NOT NULL,
    url           TEXT                                                                             NOT NULL,
    is_raw        BOOLEAN                                                                          NOT NULL,
    status        VARCHAR(64) CHECK ( status IN ('queuing', 'processing', 'completed', 'failed') ) NOT NULL DEFAULT 'queuing',
    failed_reason TEXT                                                                             NULL,
    title         TEXT                                                                             NULL,
    content       TEXT                                                                             NULL,
    token_count   INTEGER                                                                          NOT NULL DEFAULT 0,
    created_at    TIMESTAMP                                                                        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP                                                                        NOT NULL DEFAULT CURRENT_TIMESTAMP
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
DROP TABLE tasks;
DROP TABLE uploads;
DROP TABLE categories;
DROP TABLE users;
DROP FUNCTION auto_updated_at;
-- +goose StatementEnd
