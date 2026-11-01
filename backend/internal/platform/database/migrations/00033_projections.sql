-- +goose Up
-- Velocity in whole months per transition and the allowed goal stages live on the model beside the two
-- conversion rates they qualify (047-C1, 047-C2). Backfills keep every existing model meaning what it
-- meant: zero lag, every stage allowed.
ALTER TABLE models
    ADD COLUMN inq_to_mql_velocity_months  integer NOT NULL DEFAULT 0
        CHECK (inq_to_mql_velocity_months BETWEEN 0 AND 24),
    ADD COLUMN mql_to_deal_velocity_months integer NOT NULL DEFAULT 0
        CHECK (mql_to_deal_velocity_months BETWEEN 0 AND 24),
    ADD COLUMN allowed_goal_stages         text[]  NOT NULL DEFAULT '{inq,mql,revenue}'
        CHECK (cardinality(allowed_goal_stages) >= 1
               AND allowed_goal_stages <@ ARRAY['inq', 'mql', 'revenue']);

ALTER TABLE models
    ALTER COLUMN inq_to_mql_velocity_months DROP DEFAULT,
    ALTER COLUMN mql_to_deal_velocity_months DROP DEFAULT,
    ALTER COLUMN allowed_goal_stages DROP DEFAULT;

CREATE TABLE activity_categories (
    id              text        PRIMARY KEY CHECK (id ~ '^ACAT-[0-9]+$'),
    name            text        NOT NULL CHECK (name <> ''),
    is_default      boolean     NOT NULL,
    active          boolean     NOT NULL,
    version         integer     NOT NULL CHECK (version >= 1),
    created_at      timestamptz NOT NULL,
    created_by_id   text        NOT NULL,
    created_by_name text        NOT NULL,
    updated_at      timestamptz NOT NULL,
    updated_by_id   text        NOT NULL,
    updated_by_name text        NOT NULL
);

CREATE UNIQUE INDEX activity_categories_name_key ON activity_categories (lower(name COLLATE "C"));
CREATE UNIQUE INDEX activity_categories_default_key ON activity_categories (is_default) WHERE is_default;

-- The primary key IS the rule: an activity type belongs to at most one category (047-C7).
CREATE TABLE activity_category_types (
    activity_type_id text PRIMARY KEY REFERENCES activity_types (id),
    category_id      text NOT NULL REFERENCES activity_categories (id) ON DELETE CASCADE
);

CREATE INDEX activity_category_types_category_idx ON activity_category_types (category_id);

-- +goose Down
DROP TABLE activity_category_types;
DROP TABLE activity_categories;
ALTER TABLE models
    DROP COLUMN inq_to_mql_velocity_months,
    DROP COLUMN mql_to_deal_velocity_months,
    DROP COLUMN allowed_goal_stages;
