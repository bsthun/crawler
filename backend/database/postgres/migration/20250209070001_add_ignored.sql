-- +goose Up
-- +goose StatementBegin
ALTER TABLE tasks DROP CONSTRAINT tasks_status_check;
ALTER TABLE tasks ADD CONSTRAINT tasks_status_check CHECK ( status IN ('queuing', 'processing', 'completed', 'failed', 'ignored') );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tasks DROP CONSTRAINT tasks_status_check;
ALTER TABLE tasks ADD CONSTRAINT tasks_status_check CHECK ( status IN ('queuing', 'processing', 'completed', 'failed') );
-- +goose StatementEnd