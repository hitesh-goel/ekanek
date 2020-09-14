CREATE EXTENSION "uuid-ossp";

CREATE TABLE IF NOT EXISTS users
(
    uid        UUID PRIMARY KEY     DEFAULT uuid_generate_v1mc(),
    email      TEXT        NOT NULL,
    password   TEXT        NOT NULL,
    first_name TEXT        NOT NULL,
    last_name  TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (email)
);

CREATE TABLE IF NOT EXISTS assets
(
    id         UUID PRIMARY KEY     DEFAULT uuid_generate_v1mc(),
    uid        UUID        NOT NULL,
    name       TEXT        NOT NULL,
    public     BOOLEAN              DEFAULT false,
    s3_path    TEXT        NOT NULL,
    is_active  BOOLEAN              DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (uid, name),
    CONSTRAINT fk_uid
        FOREIGN KEY (uid)
            REFERENCES users (uid)
);

CREATE OR REPLACE FUNCTION updated_at_fn()
    RETURNS TRIGGER AS
$$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER users_updated_at_trigger
    BEFORE UPDATE
    ON users
    FOR EACH ROW
EXECUTE PROCEDURE updated_at_fn();

CREATE TRIGGER assets_updated_at_trigger
    BEFORE UPDATE
    ON assets
    FOR EACH ROW
EXECUTE PROCEDURE updated_at_fn();
