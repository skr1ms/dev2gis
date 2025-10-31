CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE heightmap_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    image_url VARCHAR(500) NOT NULL,
    result_url VARCHAR(500),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    width INTEGER,
    height INTEGER,
    error_message TEXT,
    processing_time FLOAT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_heightmap_jobs_user_id ON heightmap_jobs(user_id);
CREATE INDEX idx_heightmap_jobs_status ON heightmap_jobs(status);
CREATE INDEX idx_heightmap_jobs_created_at ON heightmap_jobs(created_at DESC);

CREATE TABLE batch_heightmap_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    result_url TEXT,
    orthophoto_url TEXT,
    width INTEGER,
    height INTEGER,
    image_count INTEGER NOT NULL DEFAULT 0,
    processed_count INTEGER NOT NULL DEFAULT 0,
    error_message TEXT,
    processing_time REAL,
    merge_method VARCHAR(50) NOT NULL DEFAULT 'average',
    generation_mode VARCHAR(50) NOT NULL DEFAULT 'heightmap',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE batch_images (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    batch_job_id UUID NOT NULL REFERENCES batch_heightmap_jobs(id) ON DELETE CASCADE,
    image_url TEXT NOT NULL,
    heightmap_job_id UUID REFERENCES heightmap_jobs(id) ON DELETE SET NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_batch_heightmap_jobs_user_id ON batch_heightmap_jobs(user_id);
CREATE INDEX idx_batch_heightmap_jobs_status ON batch_heightmap_jobs(status);
CREATE INDEX idx_batch_images_batch_job_id ON batch_images(batch_job_id);
CREATE INDEX idx_batch_images_status ON batch_images(status);