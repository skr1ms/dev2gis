-- name: CreateBatchHeightmapJob :one
INSERT INTO batch_heightmap_jobs (
    id, user_id, status, image_count, merge_method, generation_mode, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: CreateBatchImage :one
INSERT INTO batch_images (
    id, batch_job_id, image_url, status, created_at
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetBatchHeightmapJob :one
SELECT * FROM batch_heightmap_jobs
WHERE id = $1 LIMIT 1;

-- name: GetBatchHeightmapJobByUserID :one
SELECT * FROM batch_heightmap_jobs
WHERE id = $1 AND user_id = $2 LIMIT 1;

-- name: ListUserBatchHeightmaps :many
SELECT * FROM batch_heightmap_jobs
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetBatchImages :many
SELECT * FROM batch_images
WHERE batch_job_id = $1
ORDER BY created_at ASC;

-- name: UpdateBatchJobStatus :exec
UPDATE batch_heightmap_jobs
SET status = $2, updated_at = $3
WHERE id = $1;

-- name: UpdateBatchJobResult :exec
UPDATE batch_heightmap_jobs
SET status = $2, result_url = $3, orthophoto_url = $4, width = $5, height = $6, 
    processing_time = $7, updated_at = $8
WHERE id = $1;

-- name: UpdateBatchJobError :exec
UPDATE batch_heightmap_jobs
SET status = 'failed', error_message = $2, updated_at = $3
WHERE id = $1;

-- name: UpdateBatchJobProgress :exec
UPDATE batch_heightmap_jobs
SET processed_count = $2, updated_at = $3
WHERE id = $1;

-- name: UpdateBatchImageStatus :exec
UPDATE batch_images
SET status = $2, heightmap_job_id = $3
WHERE id = $1;

