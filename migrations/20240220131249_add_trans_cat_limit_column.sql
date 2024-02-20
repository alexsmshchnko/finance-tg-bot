-- +goose Up
-- +goose StatementBegin
ALTER TABLE trans_category ADD trans_limit INT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE trans_category DROP COLUMN trans_limit;
-- +goose StatementEnd
