-- +goose Up
-- +goose StatementBegin
CREATE TABLE
    IF NOT EXISTS viewer_bufo (
        name text NOT NULL PRIMARY KEY,
        created timestamptz NOT NULL
    );

CREATE TABLE
    IF NOT EXISTS viewer_bufovote (
        id BIGSERIAL NOT NULL PRIMARY KEY,
        "value" integer NOT NULL,
        created timestamptz NOT NULL,
        bufo_id text NOT NULL REFERENCES viewer_bufo(name)
    );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE viewer_bufo;

DROP TABLE viewer_bufovote;
-- +goose StatementEnd
