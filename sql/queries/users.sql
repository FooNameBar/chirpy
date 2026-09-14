-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING id, created_at, updated_at, email;

-- name: ResetUsers :one
WITH deleted AS (
    DELETE FROM users
    RETURNING *
)
SELECT COUNT(*) AS deleted_count
FROM deleted;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email=$1;
