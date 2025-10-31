DROP INDEX IF EXISTS idx_batch_images_status;
DROP INDEX IF EXISTS idx_batch_images_batch_job_id;
DROP INDEX IF EXISTS idx_batch_heightmap_jobs_status;
DROP INDEX IF EXISTS idx_batch_heightmap_jobs_user_id;
DROP TABLE IF EXISTS batch_images;
DROP TABLE IF EXISTS batch_heightmap_jobs;

DROP INDEX IF EXISTS idx_heightmap_jobs_created_at;
DROP INDEX IF EXISTS idx_heightmap_jobs_status;
DROP INDEX IF EXISTS idx_heightmap_jobs_user_id;
DROP TABLE IF EXISTS heightmap_jobs;

DROP TABLE IF EXISTS users;