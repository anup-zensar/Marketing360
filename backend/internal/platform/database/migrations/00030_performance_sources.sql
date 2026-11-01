-- +goose Up
-- Feature 046: where performance data comes from, and what each column of its file means.
CREATE TABLE performance_sources (
    id              text        PRIMARY KEY CHECK (id ~ '^PS-[0-9]+$'),
    name            text        NOT NULL CHECK (char_length(name) BETWEEN 1 AND 120),
    description     text        NOT NULL CHECK (char_length(description) <= 2000),
    owner_id        text        NOT NULL REFERENCES people (id) DEFERRABLE INITIALLY DEFERRED,
    category        text        NOT NULL
                                CHECK (category IN ('webAnalytics', 'marketingAutomation', 'crm', 'leadProgression')),
    status          text        NOT NULL CHECK (status IN ('active', 'inactive')),
    delivery_route  text        NOT NULL CHECK (delivery_route IN ('upload', 'simulated')),
    cadence         text        NOT NULL CHECK (cadence IN ('daily', 'weekly', 'monthly', 'onDemand')),
    integration_id  text        NOT NULL,
    version         integer     NOT NULL CHECK (version >= 1),
    created_at      timestamptz NOT NULL,
    created_by_id   text        NOT NULL,
    created_by_name text        NOT NULL,
    updated_at      timestamptz NOT NULL,
    updated_by_id   text        NOT NULL,
    updated_by_name text        NOT NULL
);

CREATE UNIQUE INDEX performance_sources_name_idx ON performance_sources (lower(name COLLATE "C"));
CREATE INDEX performance_sources_status_idx ON performance_sources (status);
CREATE UNIQUE INDEX performance_sources_integration_id_idx
    ON performance_sources (integration_id) WHERE integration_id <> '';

CREATE TABLE performance_source_mappings (
    id              text        PRIMARY KEY CHECK (id ~ '^PSM-[0-9]+$'),
    source_id       text        NOT NULL REFERENCES performance_sources (id) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED,
    column_name     text        NOT NULL CHECK (char_length(column_name) BETWEEN 1 AND 120),
    meaning         text        NOT NULL
                                CHECK (meaning IN ('key', 'reference', 'factDate', 'stage', 'revenue',
                                                   'currency', 'webMetric', 'journey')),
    key_type        text        NOT NULL
                                CHECK (key_type IN ('', 'activityId', 'mediaCode', 'integrationId', 'externalName')),
    stage           text        NOT NULL CHECK (stage IN ('', 'inq', 'mql', 'revenue')),
    web_metric      text        NOT NULL
                                CHECK (web_metric IN ('', 'sessions', 'visits', 'engagementTouches')),
    model_id        text        REFERENCES models (id) DEFERRABLE INITIALLY DEFERRED,
    version         integer     NOT NULL CHECK (version >= 1),
    created_at      timestamptz NOT NULL,
    created_by_id   text        NOT NULL,
    created_by_name text        NOT NULL,
    updated_at      timestamptz NOT NULL,
    updated_by_id   text        NOT NULL,
    updated_by_name text        NOT NULL,
    CHECK ((meaning = 'key') = (key_type <> '')),
    CHECK ((meaning = 'stage') = (stage <> '')),
    CHECK ((meaning = 'webMetric') = (web_metric <> '')),
    CHECK ((meaning IN ('stage', 'revenue')) = (model_id IS NOT NULL))
);

CREATE UNIQUE INDEX performance_source_mappings_column_idx
    ON performance_source_mappings (source_id, lower(column_name COLLATE "C"));
CREATE UNIQUE INDEX performance_source_mappings_key_idx
    ON performance_source_mappings (source_id, key_type) WHERE meaning = 'key';

-- One ACTIVE source per model stage and per model revenue spans two tables, so the rule lives in the
-- service rather than a constraint (ADR-0021); these indexes only keep its check fast.
CREATE INDEX performance_source_mappings_model_stage_idx
    ON performance_source_mappings (model_id, stage) WHERE meaning = 'stage';
CREATE INDEX performance_source_mappings_model_revenue_idx
    ON performance_source_mappings (model_id) WHERE meaning = 'revenue';

-- Table A gains D-367's two 046 keys; reports.view stays listed whichever of 039 and this runs first.
ALTER TABLE roles
    DROP CONSTRAINT roles_permissions_check,
    ADD CONSTRAINT roles_permissions_check CHECK (cardinality(permissions) >= 1 AND permissions <@ ARRAY[
        'plans.view', 'plans.create', 'plans.edit', 'plans.publish', 'plans.delete',
        'campaigns.edit', 'campaigns.delete', 'activities.edit', 'activities.delete',
        'activities.submit', 'data.export', 'reports.view', 'finance.view', 'finance.link', 'finance.import',
        'performance.import', 'performance.link',
        'purchaseRequests.edit', 'purchaseRequests.delete', 'purchaseRequests.submit', 'purchaseRequests.list',
        'approvals.decide', 'approvals.decideOnBehalf', 'globalBudgets.view',
        'globalBudgets.manage', 'settings.manage', 'exchangeRates.manage', 'workflows.manage',
        'people.manage', 'access.manage', 'activityLog.viewAll', 'recycleBin.purge']);
