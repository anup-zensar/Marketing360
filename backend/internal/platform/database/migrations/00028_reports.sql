-- +goose Up
-- Feature 039: saved report views, and reports.view for every role that may view plans, kept in Table A order (D-367).
CREATE TABLE report_views (
    id              text        PRIMARY KEY CHECK (id ~ '^RV-[0-9]+$'),
    person_id       text        NOT NULL REFERENCES people (id) DEFERRABLE INITIALLY DEFERRED,
    report          text        NOT NULL CHECK (report IN ('planBudget', 'approvals', 'purchaseRequests', 'planDetails',
        'gsapLink', 'userAccess', 'summary', 'revenue', 'funnel')),
    name            text        NOT NULL CHECK (name <> '' AND char_length(name) <= 60),
    is_default      boolean     NOT NULL,
    query           text        NOT NULL CHECK (char_length(query) <= 4000),
    version         integer     NOT NULL CHECK (version >= 1),
    created_at      timestamptz NOT NULL,
    created_by_id   text        NOT NULL,
    created_by_name text        NOT NULL,
    updated_at      timestamptz NOT NULL,
    updated_by_id   text        NOT NULL,
    updated_by_name text        NOT NULL
);
CREATE UNIQUE INDEX report_views_name_key ON report_views (person_id, report, lower(name COLLATE "C"));
CREATE UNIQUE INDEX report_views_default_key ON report_views (person_id, report) WHERE is_default;

ALTER TABLE roles
    DROP CONSTRAINT roles_permissions_check,
    ADD CONSTRAINT roles_permissions_check CHECK (cardinality(permissions) >= 1 AND permissions <@ ARRAY[
        'plans.view', 'plans.create', 'plans.edit', 'plans.publish', 'plans.delete',
        'campaigns.edit', 'campaigns.delete', 'activities.edit', 'activities.delete',
        'activities.submit', 'data.export', 'reports.view', 'finance.view', 'finance.link', 'finance.import',
        'purchaseRequests.edit', 'purchaseRequests.delete', 'purchaseRequests.submit', 'purchaseRequests.list',
        'approvals.decide', 'approvals.decideOnBehalf', 'globalBudgets.view',
        'globalBudgets.manage', 'settings.manage', 'exchangeRates.manage', 'workflows.manage',
        'people.manage', 'access.manage', 'activityLog.viewAll', 'recycleBin.purge']);

UPDATE roles
SET permissions = ARRAY(
        SELECT key
        FROM unnest(ARRAY[
            'plans.view', 'plans.create', 'plans.edit', 'plans.publish', 'plans.delete',
            'campaigns.edit', 'campaigns.delete', 'activities.edit', 'activities.delete',
            'activities.submit', 'data.export', 'reports.view', 'finance.view', 'finance.link', 'finance.import',
            'purchaseRequests.edit', 'purchaseRequests.delete', 'purchaseRequests.submit', 'purchaseRequests.list',
            'approvals.decide', 'approvals.decideOnBehalf', 'globalBudgets.view',
            'globalBudgets.manage', 'settings.manage', 'exchangeRates.manage', 'workflows.manage',
            'people.manage', 'access.manage', 'activityLog.viewAll', 'recycleBin.purge']) WITH ORDINALITY AS table_a (key, position)
        WHERE key = ANY (permissions) OR key = 'reports.view'
        ORDER BY position),
    version = version + 1, updated_at = now(), updated_by_id = 'system', updated_by_name = 'SP&A One'
WHERE 'plans.view' = ANY (permissions) AND NOT 'reports.view' = ANY (permissions);
