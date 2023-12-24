-- name: CreateRecord :one
INSERT INTO Records
(zone, content, name, is_wildcard, type, comment)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ReadRecord :one
SELECT * FROM Records
WHERE id = $1;

-- name: ResolveRecord :many
SELECT * FROM Records
WHERE name = $1 and type = $2 and is_wildcard = false;

-- name: ResolveWildcardRecord :many
SELECT * FROM Records
WHERE name = ANY(sqlc.arg(names)::text[]) and type = $1 and is_wildcard = true;
