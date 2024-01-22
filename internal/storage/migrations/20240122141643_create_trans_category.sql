-- +goose Up
-- +goose StatementBegin
CREATE TABLE trans_category
(
    id              SERIAL PRIMARY KEY,
    trans_cat       VARCHAR(255) NOT NULL,
    direction       SMALLINT     NOT NULL DEFAULT 0,
    client_id       VARCHAR(255),
    active          BOOL         NOT NULL DEFAULT true
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS trans_category;
-- +goose StatementEnd
