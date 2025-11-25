CREATE TABLE users
(
    id INTEGER PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    pass_hash BLOB NOT NULL
);

CREATE INDEX idx_email ON users(email);

CREATE TABLE apps
(
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    secret TEXT NOT NULL UNIQUE
);