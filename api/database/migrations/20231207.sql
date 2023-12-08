-- +migrate Up
CREATE TABLE Records (
	id SERIAL PRIMARY KEY,

	-- Record
	zone TEXT NOT NULL,
	content TEXT NOT NULL,

	-- Metadata
	name TEXT NOT NULL,
	is_wildcard BOOLEAN NOT NULL,
	type INTEGER NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	modified_on TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	comment TEXT NOT NULL
);

-- +migrate Down
DROP TABLE Records;
