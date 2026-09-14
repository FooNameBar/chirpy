-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1
)
RETURNING *;

-- name: ResetUsers :one
WITH deleted AS (
    DELETE FROM users
    RETURNING *
)
SELECT COUNT(*) AS deleted_count
FROM deleted;
