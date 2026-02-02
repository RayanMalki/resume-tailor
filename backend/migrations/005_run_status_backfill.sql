-- +goose Up
-- +goose StatementBegin

UPDATE runs SET status = 'queued' WHERE status = 'created';
UPDATE runs SET status = 'running' WHERE status = 'processing';
UPDATE runs SET status = 'succeeded' WHERE status = 'completed';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

UPDATE runs SET status = 'completed' WHERE status = 'succeeded';
UPDATE runs SET status = 'processing' WHERE status = 'running';
UPDATE runs SET status = 'created' WHERE status = 'queued';

-- +goose StatementEnd
