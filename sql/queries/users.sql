-- The :one at the end of the query name tells SQLC that we expect to get back a single row 
-- (the created user)
-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email)
VALUES (
    --  you can use gen_random_uuid() to generate a new UUID:
    gen_random_uuid(),
    --  set to the current timestamp. In Postgres, you can use NOW() to get the current timestamp:
    NOW(),
    NOW(),
    -- The email should be passed in by our application. Use $1 to represent the first parameter 
    -- passed into the query:
    $1
)
RETURNING *;