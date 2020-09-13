CREATE EXTENSION "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    uid UUID PRIMARY KEY DEFAULT uuid_generate_v1mc(),
    email TEXT NOT NULL,
    password TEXT NOT NULL,
    first_name TEXT NOT NULL,
    last_name TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (email)
);

CREATE OR REPLACE FUNCTION updated_at_fn()
    RETURNS TRIGGER AS $$
    BEGIN
        NEW.updated_at =  NOW();
        RETURN NEW;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER updated_at_trigger
    BEFORE UPDATE
    ON users
    FOR EACH ROW
    EXECUTE PROCEDURE updated_at_fn();
