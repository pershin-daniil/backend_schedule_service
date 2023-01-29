-- noinspection SqlNoDataSourceInspectionForFile

-- +migrate Up

CREATE TABLE users
(
    id serial PRIMARY KEY,
    last_name varchar NOT NULL,
    first_name varchar NOT NULL
);

CREATE TABLE meetings
(
    id serial PRIMARY KEY,
    manager int NOT NULL REFERENCES users (id),
    start_at int NOT NULL,
    end_at int NOT NULL,
    client int NOT NULL REFERENCES users (id)
);

-- +migrate Down

DROP TABLE users;