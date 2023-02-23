-- noinspection SqlNoDataSourceInspectionForFile

-- +migrate Up

CREATE TABLE users_history
(
    id serial PRIMARY KEY,
    user_id int NOT NULL,
    last_name varchar NOT NULL,
    first_name varchar NOT NULL,
    phone varchar NOT NULL,
    email varchar,
    event_time timestamptz NOT NULL DEFAULT NOW(),
    created_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE TABLE meetings_history
(
    id serial PRIMARY KEY,
    meetings_id int NOT NULL,
    manager int NOT NULL,
    start_at timestamptz NOT NULL,
    end_at timestamptz NOT NULL,
    client int,
    event_time timestamptz NOT NULL DEFAULT NOW(),
    created_at timestamptz NOT NULL DEFAULT NOW()
);

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION users_history()
    RETURNS TRIGGER AS
$$
BEGIN
    INSERT INTO users_history (user_id, last_name, first_name, phone, email, event_time, created_at)
    VALUES (NEW.id, NEW.last_name, NEW.first_name, NEW.phone, NEW.email, NOW(), NEW.created_at);
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd


-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION meetings_history()
    RETURNS TRIGGER AS
$$
BEGIN
    INSERT INTO meetings_history (meetings_id, manager, start_at, end_at, client, event_time, created_at)
    VALUES (NEW.id, NEW.manager, NEW.start_at, NEW.end_at, NEW.client, NOW(), NEW.created_at);
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd

DROP TRIGGER IF EXISTS users_history_update ON users;
CREATE TRIGGER users_history_update
    AFTER UPDATE ON users
    FOR EACH ROW
EXECUTE PROCEDURE users_history();

DROP TRIGGER IF EXISTS users_history_delete ON users;
CREATE TRIGGER users_history_delete
    AFTER DELETE ON users
    FOR EACH ROW
EXECUTE PROCEDURE users_history();

DROP TRIGGER IF EXISTS users_history_insert ON users;
CREATE TRIGGER users_history_insert
    AFTER INSERT ON users
    FOR EACH ROW
EXECUTE PROCEDURE users_history();

DROP TRIGGER IF EXISTS meetings_history_update ON meetings;
CREATE TRIGGER meetings_history_update
    AFTER UPDATE ON meetings
    FOR EACH ROW
EXECUTE PROCEDURE meetings_history();


DROP TRIGGER IF EXISTS meetings_history_insert ON meetings;
CREATE TRIGGER meetings_history_insert
    AFTER INSERT ON meetings
    FOR EACH ROW
EXECUTE PROCEDURE meetings_history();

-- +migrate Down

DROP TRIGGER users_history_update ON users;
DROP TRIGGER users_history_delete ON users;
DROP TRIGGER users_history_insert ON users;
DROP FUNCTION users_history();
DROP TABLE users_history;

DROP TRIGGER meetings_history_update ON meetings;
DROP TRIGGER meetings_history_insert ON meetings;
DROP FUNCTION meetings_history();
DROP TABLE meetings_history;
