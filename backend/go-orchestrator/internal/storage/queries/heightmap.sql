-- name: CreateHeightmapJob :one
INSERT INTO heightmap_jobs (
    id, user_id, image_url, status, created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetHeightmapJob :one
SELECT * FROM heightmap_jobs WHERE id = $1;

-- name: GetHeightmapJobByUserID :one
SELECT * FROM heightmap_jobs WHERE id = $1 AND user_id = $2;

-- name: UpdateJobStatus :exec
UPDATE heightmap_jobs
SET status = $2, updated_at = $3
WHERE id = $1;

-- name: UpdateJobResult :exec
UPDATE heightmap_jobs
SET 
    status = $2,
    result_url = $3,
    width = $4,
    height = $5,
    processing_time = $6,
    updated_at = $7
WHERE id = $1;

-- name: UpdateJobError :exec
UPDATE heightmap_jobs
SET 
    status = 'failed',
    error_message = $2,
    updated_at = $3
WHERE id = $1;

-- name: ListUserHeightmaps :many
SELECT * FROM heightmap_jobs
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountUserHeightmaps :one
SELECT COUNT(*) FROM heightmap_jobs
WHERE user_id = $1;

-- name: ListHeightmapsByStatus :many
SELECT * FROM heightmap_jobs
WHERE status = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: DeleteHeightmapJob :exec
DELETE FROM heightmap_jobs WHERE id = $1;

