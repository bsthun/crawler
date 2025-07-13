-- +goose Up
-- +goose StatementBegin
ALTER TABLE tasks ADD COLUMN revised_task_id BIGINT NULL;
ALTER TABLE tasks ADD CONSTRAINT fk_tasks_revised_task_id FOREIGN KEY (revised_task_id) REFERENCES tasks(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tasks DROP CONSTRAINT fk_tasks_revised_task_id;
ALTER TABLE tasks DROP COLUMN revised_task_id;
-- +goose StatementEnd