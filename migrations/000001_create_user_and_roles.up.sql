CREATE TABLE IF NOT EXISTS roles(
    id serial primary key,
    name text not null
);

INSERT INTO roles (name) VALUES ('DSVR'), ('ADMIN'), ('MODER'), ('USER');

CREATE TABLE IF NOT EXISTS users(
    id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    pass_hash BYTEA NOT NULL,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    created_at timestamp DEFAULT CURRENT_TIMESTAMP,
    activated boolean not null default false,
    locked boolean not null default false,
    barcode VARCHAR(255) NOT NULL UNIQUE DEFAULT '',
    phone_number VARCHAR(255) NOT NULL DEFAULT '',
    major TEXT NOT NULL DEFAULT '',
    group_name TEXT NOT NULL DEFAULT '',
    year INTEGER NOT NULL DEFAULT 0,
    role_id INTEGER NOT NULL REFERENCES roles(id) DEFAULT 4
);