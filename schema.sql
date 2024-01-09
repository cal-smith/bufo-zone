PRAGMA journal_mode = WAL;

CREATE TABLE
    IF NOT EXISTS viewer_bufo (
        "name" text NOT NULL PRIMARY KEY,
        "created" datetime NOT NULL
    );

CREATE TABLE
    IF NOT EXISTS viewer_bufovote (
        "id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
        "value" integer NOT NULL,
        "created" datetime NOT NULL,
        "bufo_id" text NOT NULL REFERENCES "viewer_bufo" ("name") DEFERRABLE INITIALLY DEFERRED
    );