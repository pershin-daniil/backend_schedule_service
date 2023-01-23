-- noinspection SqlNoDataSourceInspectionForFile

-- +migrate Up

CREATE TABLE users
(
    user_id serial PRIMARY KEY,
    last_name varchar NOT NULL,
    first_name varchar NOT NULL
);

CREATE TABLE meetings
(
    id serial PRIMARY KEY,
    creator integer NOT NULL REFERENCES users (user_id)
    -- ....
);

-- +migrate Down

DROP TABLE users;