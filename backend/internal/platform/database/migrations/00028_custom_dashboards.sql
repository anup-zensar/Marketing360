-- +goose Up
CREATE TABLE custom_dashboards (
    id               text        PRIMARY KEY CHECK (id ~ '^DB-[0-9]+$'),
    name             text        NOT NULL CHECK (char_length(name) BETWEEN 1 AND 80 AND name = btrim(name)),
    description      text        NOT NULL CHECK (char_length(description) <= 500),
    owner_id         text        NOT NULL REFERENCES people (id) DEFERRABLE INITIALLY DEFERRED,
    filter_date_from date,
    filter_date_to   date,
    filter_plan_ids  text[]      NOT NULL CHECK (cardinality(filter_plan_ids) <= 20),
    filter_team_ids  text[]      NOT NULL CHECK (cardinality(filter_team_ids) <= 20),
    version          integer     NOT NULL CHECK (version >= 1),
    created_at       timestamptz NOT NULL,
    created_by_id    text        NOT NULL,
    created_by_name  text        NOT NULL,
    updated_at       timestamptz NOT NULL,
    updated_by_id    text        NOT NULL,
    updated_by_name  text        NOT NULL,
    CHECK (filter_date_from IS NULL OR filter_date_to IS NULL OR filter_date_from <= filter_date_to)
);
CREATE UNIQUE INDEX custom_dashboards_owner_name_key ON custom_dashboards (owner_id, lower(name));

CREATE TABLE custom_dashboard_widgets (
    dashboard_id text    NOT NULL REFERENCES custom_dashboards (id) ON DELETE CASCADE,
    id           text    NOT NULL CHECK (id ~ '^w[1-9][0-9]{0,2}$'),
    position     integer NOT NULL CHECK (position BETWEEN 0 AND 29),
    type         text    NOT NULL CHECK (type IN ('bar', 'line', 'area', 'combo', 'pie', 'donut', 'scatter', 'funnel',
                                                  'gauge', 'heatmap', 'treemap', 'radar', 'kpi', 'table', 'text')),
    title        text    NOT NULL CHECK (char_length(title) <= 80),
    subtitle     text    NOT NULL CHECK (char_length(subtitle) <= 120),
    grid_x       integer NOT NULL CHECK (grid_x BETWEEN 0 AND 11),
    grid_y       integer NOT NULL CHECK (grid_y BETWEEN 0 AND 199),
    grid_w       integer NOT NULL CHECK (grid_w BETWEEN 2 AND 12),
    grid_h       integer NOT NULL CHECK (grid_h BETWEEN 2 AND 12),
    source       text    CHECK (source IS NULL OR source <> ''),
    query        jsonb   CHECK (query IS NULL OR jsonb_typeof(query) = 'object'),
    display      jsonb   NOT NULL CHECK (jsonb_typeof(display) = 'object'),
    note         text    NOT NULL CHECK (char_length(note) <= 2000),
    PRIMARY KEY (dashboard_id, id),
    UNIQUE (dashboard_id, position),
    CHECK (grid_x + grid_w <= 12 AND grid_y + grid_h <= 200),
    CHECK ((type = 'text') = (query IS NULL) AND (query IS NULL) = (source IS NULL))
);
CREATE INDEX custom_dashboard_widgets_source_idx ON custom_dashboard_widgets (source) WHERE source LIKE 'DS-%';

CREATE TABLE custom_dashboard_shares (
    dashboard_id text NOT NULL REFERENCES custom_dashboards (id) ON DELETE CASCADE,
    team_id      text NOT NULL REFERENCES teams (id) DEFERRABLE INITIALLY DEFERRED,
    PRIMARY KEY (dashboard_id, team_id)
);
CREATE INDEX custom_dashboard_shares_team_idx ON custom_dashboard_shares (team_id);

ALTER TABLE roles
    DROP CONSTRAINT roles_permissions_check,
    ADD CONSTRAINT roles_permissions_check CHECK (cardinality(permissions) >= 1 AND permissions <@ ARRAY[
        'plans.view', 'plans.create', 'plans.edit', 'plans.publish', 'plans.delete',
        'campaigns.edit', 'campaigns.delete', 'activities.edit', 'activities.delete',
        'activities.submit', 'data.export', 'finance.view', 'finance.link', 'finance.import',
        'purchaseRequests.edit', 'purchaseRequests.delete', 'purchaseRequests.submit', 'purchaseRequests.list',
        'approvals.decide', 'approvals.decideOnBehalf', 'globalBudgets.view',
        'globalBudgets.manage', 'dashboards.view', 'dashboards.build', 'settings.manage', 'exchangeRates.manage',
        'workflows.manage', 'people.manage', 'access.manage', 'activityLog.viewAll', 'recycleBin.purge']);

-- Feature 051: the built-in roles see and build dashboards (D-374).
UPDATE roles SET permissions = array_append(permissions, 'dashboards.view'), version = version + 1,
    updated_at = now(), updated_by_id = 'system', updated_by_name = 'SP&A One'
WHERE built_in <> '' AND NOT 'dashboards.view' = ANY (permissions);
UPDATE roles SET permissions = array_append(permissions, 'dashboards.build'), version = version + 1,
    updated_at = now(), updated_by_id = 'system', updated_by_name = 'SP&A One'
WHERE built_in IN ('administrator', 'globalPlanner', 'planner') AND NOT 'dashboards.build' = ANY (permissions);
