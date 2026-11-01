-- +goose Up
-- Feature 051: datasets uploaded from CSV or .xlsx files, typed columns, one JSON array per row, shared with teams.
CREATE TABLE datasets (
    id              text        PRIMARY KEY CHECK (id ~ '^DS-[0-9]+$'),
    name            text        NOT NULL CHECK (char_length(name) BETWEEN 1 AND 80 AND name = btrim(name)),
    description     text        NOT NULL CHECK (char_length(description) <= 500),
    owner_id        text        NOT NULL REFERENCES people (id) DEFERRABLE INITIALLY DEFERRED,
    file_name       text        NOT NULL CHECK (char_length(file_name) BETWEEN 1 AND 255),
    sheet_name      text        NOT NULL CHECK (char_length(sheet_name) <= 255),
    row_count       integer     NOT NULL CHECK (row_count BETWEEN 0 AND 10000),
    version         integer     NOT NULL CHECK (version >= 1),
    created_at      timestamptz NOT NULL,
    created_by_id   text        NOT NULL,
    created_by_name text        NOT NULL,
    updated_at      timestamptz NOT NULL,
    updated_by_id   text        NOT NULL,
    updated_by_name text        NOT NULL
);
CREATE UNIQUE INDEX datasets_owner_name_key ON datasets (owner_id, lower(name));

CREATE TABLE dataset_columns (
    dataset_id text    NOT NULL REFERENCES datasets (id) ON DELETE CASCADE,
    position   integer NOT NULL CHECK (position BETWEEN 0 AND 49),
    key        text    NOT NULL CHECK (key = 'c' || (position + 1)),
    label      text    NOT NULL CHECK (char_length(label) BETWEEN 1 AND 60),
    kind       text    NOT NULL CHECK (kind IN ('text', 'number', 'date', 'money')),
    currency   text    CHECK (currency ~ '^[A-Z]{3}$'),
    PRIMARY KEY (dataset_id, position),
    CHECK ((kind = 'money') = (currency IS NOT NULL))
);
CREATE UNIQUE INDEX dataset_columns_label_key ON dataset_columns (dataset_id, lower(label));

CREATE TABLE dataset_rows (
    dataset_id text    NOT NULL REFERENCES datasets (id) ON DELETE CASCADE,
    row_number integer NOT NULL CHECK (row_number BETWEEN 1 AND 10000),
    cells      jsonb   NOT NULL CHECK (jsonb_typeof(cells) = 'array' AND jsonb_array_length(cells) <= 50),
    PRIMARY KEY (dataset_id, row_number)
);

CREATE TABLE dataset_shares (
    dataset_id text NOT NULL REFERENCES datasets (id) ON DELETE CASCADE,
    team_id    text NOT NULL REFERENCES teams (id) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED,
    PRIMARY KEY (dataset_id, team_id)
);
CREATE INDEX dataset_shares_team_idx ON dataset_shares (team_id);
