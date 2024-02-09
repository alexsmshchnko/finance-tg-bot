-- +goose Up
-- +goose StatementBegin
CREATE TABLE client
(
    id                    SERIAL PRIMARY KEY,
	username              VARCHAR(255) NOT NULL,
	first_login_date      DATE DEFAULT NOW(),
	is_active             BOOL         NOT NULL,
	external_system_name  VARCHAR(255),
	external_system_token VARCHAR(255)
);
CREATE UNIQUE INDEX client_username_idx ON client (username);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS client;
-- +goose StatementEnd
