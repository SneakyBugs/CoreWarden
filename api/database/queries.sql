-- name: CreateRecord :one
INSERT INTO Records
(zone, content, name, is_wildcard, type, comment)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: AnyRecordsExistAtNode :one
SELECT EXISTS(
  SELECT 1 FROM Records
  WHERE zone = $1 and name = $2
);

-- name: CNAMERecordExistsAtNode :one
SELECT EXISTS(
  SELECT 1 FROM Records
  WHERE zone = $1 and name = $2 and type = 5
);

-- name: ReadRecord :one
SELECT * FROM Records
WHERE id = $1;

-- name: UpdateRecord :one
UPDATE Records
SET zone = $1, content = $2, name = $3, is_wildcard = $4, type = $5, comment = $6, modified_on = NOW()
where id = $7
RETURNING *;

-- name: DeleteRecord :one
DELETE FROM Records
WHERE id = $1
RETURNING *;

-- name: ListRecords :many
SELECT * FROM Records
WHERE zone = $1;

-- name: ResolveRecord :many
SELECT * FROM Records
WHERE name = $1 and (type = $2 or type = 5) and is_wildcard = false;

-- name: ResolveWildcardRecord :many
SELECT * FROM Records
WHERE name = ANY(sqlc.arg(names)::text[]) and type = $1 and is_wildcard = true;
