-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password, is_chirpy_red)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2,
    FALSE
)
RETURNING id, created_at, updated_at, email, is_chirpy_red;

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

-- name: UpdateEmailPassword :one
UPDATE users
SET email=$1,
hashed_password=$2
WHERE id=$3
RETURNING id, created_at, updated_at, email, is_chirpy_red;

-- name: UpdateChirpyRed :one
UPDATE users
SET updated_at=NOW(),
is_chirpy_red=TRUE
WHERE id=$1
RETURNING id, created_at, updated_at, email, is_chirpy_red;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id=$1;
