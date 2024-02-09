-- +goose Up
-- +goose StatementBegin
CREATE TABLE document
(
    id              SERIAL PRIMARY KEY,
    trans_date      TIMESTAMP    NOT NULL DEFAULT NOW(),
    trans_cat       VARCHAR(255) NOT NULL,
    trans_amount    INT          NOT NULL,
    comment         VARCHAR(255),
    tg_msg_id       VARCHAR(255),
    client_id       VARCHAR(255)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS document;
-- +goose StatementEnd
