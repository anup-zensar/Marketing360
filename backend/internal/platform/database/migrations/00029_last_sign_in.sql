-- +goose Up
-- Feature 045: each person's latest successful sign-in (UTC), written by sign-in only and kept apart from the person,
-- so their version and "last changed" never move (045-C4); sign_in_recording keeps when recording started here.
CREATE TABLE sign_ins (
    person_id         text        PRIMARY KEY REFERENCES people (id) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED,
    last_signed_in_at timestamptz NOT NULL
);

CREATE TABLE sign_in_recording (
    id    boolean     PRIMARY KEY DEFAULT true CHECK (id),
    since timestamptz NOT NULL
);
INSERT INTO sign_in_recording (since) VALUES (now());

-- +goose Down
DROP TABLE sign_in_recording;
DROP TABLE sign_ins;
