-- +goose Up
-- Feature 050: the designated attribution model, one row, versioned like Settings › Overspend control.
CREATE TABLE attribution_settings (
    id              text        PRIMARY KEY CHECK (id = 'attribution'),
    model           text        NOT NULL CHECK (model IN ('firstTouch', 'lastTouch', 'equallyDistributed',
        'linearTimeDecay', 'uShaped')),
    version         integer     NOT NULL CHECK (version >= 1),
    created_at      timestamptz NOT NULL,
    created_by_id   text        NOT NULL,
    created_by_name text        NOT NULL,
    updated_at      timestamptz NOT NULL,
    updated_by_id   text        NOT NULL,
    updated_by_name text        NOT NULL
);
