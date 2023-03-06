-- noinspection SqlNoDataSourceInspectionForFile

-- +migrate Up

CREATE TABLE users
(
    id            serial PRIMARY KEY,
    last_name     varchar     NOT NULL,
    first_name    varchar     NOT NULL,
    phone         varchar     NOT NULL UNIQUE,
    email         varchar,
    password_hash varchar     NOT NULL,
    role          varchar     NOT NULL DEFAULT 'client',
    deleted       bool        NOT NULL DEFAULT FALSE,
    updated_at    timestamptz NOT NULL DEFAULT NOW(),
    created_at    timestamptz NOT NULL DEFAULT NOW()
);

CREATE INDEX users_phone_idx ON users (phone);

CREATE TABLE meetings
(
    id         serial PRIMARY KEY,
    manager    int         NOT NULL REFERENCES users (id),
    start_at   timestamptz NOT NULL,
    end_at     timestamptz NOT NULL,
    client     int         NOT NULL REFERENCES users (id),
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    created_at timestamptz NOT NULL DEFAULT NOW()
);

-- +migrate Down

DROP INDEX users_phone_idx;
DROP TABLE users;
DROP TABLE meetings;