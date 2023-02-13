-- noinspection SqlNoDataSourceInspectionForFile

-- +migrate Up

CREATE TABLE users
(
    id serial PRIMARY KEY,
    last_name varchar NOT NULL,
    first_name varchar NOT NULL,
    phone varchar NOT NULL,
    email varchar,
    deleted bool NOT NULL DEFAULT false,
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    created_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE TABLE meetings
(
    id serial PRIMARY KEY,
    manager int NOT NULL REFERENCES users (id),
    start_at timestamptz NOT NULL,
    end_at timestamptz NOT NULL,
    client int NOT NULL REFERENCES users (id),
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    created_at timestamptz NOT NULL DEFAULT NOW()
);

-- +migrate Down

DROP TABLE users;
DROP TABLE meetings;
