-- name: RecordSignIn :exec
INSERT INTO sign_ins (person_id, last_signed_in_at) VALUES (sqlc.arg(person_id), sqlc.arg(at))
ON CONFLICT (person_id) DO UPDATE SET last_signed_in_at = excluded.last_signed_in_at;

-- name: ListLastSignIns :many
SELECT person_id, last_signed_in_at FROM sign_ins ORDER BY length(person_id), person_id COLLATE "C";

-- name: GetSignInsRecordedSince :one
SELECT since FROM sign_in_recording;
