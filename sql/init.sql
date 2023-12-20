CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY,
    username TEXT NOT NULL,
    password BLOB NOT NULL,
    joined TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS userdata (
    id INTEGER PRIMARY KEY,
    cash INTEGER NOT NULL,
    admin BOOLEAN NOT NULL
);

CREATE TABLE IF NOT EXISTS tokens (
    id INTEGER PRIMARY KEY,
    token TEXT NOT NULL
);