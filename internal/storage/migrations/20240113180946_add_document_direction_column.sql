-- +goose Up
-- +goose StatementBegin
ALTER TABLE document ADD direction smallint NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE document DROP COLUMN direction;
-- +goose StatementEnd
