-- +goose Up
-- Feature 045: each person's latest sign-in apart from the person (045-C4), and when recording started (R2 addendum).
CREATE TABLE sign_ins (
    person_id         text        PRIMARY KEY REFERENCES people (id) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED,
    last_signed_in_at timestamptz NOT NULL
);

CREATE TABLE sign_in_recording (
    id    boolean     PRIMARY KEY DEFAULT true CHECK (id),
    since timestamptz NOT NULL
);
INSERT INTO sign_in_recording (since) VALUES (now());
