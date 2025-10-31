package rabbitmq

import "time"

type HeightmapTask struct {
	JobID          string    `json:"job_id"`
	UserID         string    `json:"user_id"`
	ImageURL       string    `json:"image_url"`
	OutputBucket   string    `json:"output_bucket"`
	MinioEndpoint  string    `json:"minio_endpoint"`
	MinioPublicURL string    `json:"minio_public_url"`
	MinioAccessKey string    `json:"minio_access_key"`
	MinioSecretKey string    `json:"minio_secret_key"`
	MinioUseSSL    bool      `json:"minio_use_ssl"`
	CreatedAt      time.Time `json:"created_at"`
	Priority       int       `json:"priority"`
}

type BatchHeightmapTask struct {
	BatchJobID     string    `json:"batch_job_id"`
	UserID         string    `json:"user_id"`
	ImageURLs      []string  `json:"image_urls"`
	OutputBucket   string    `json:"output_bucket"`
	MinioEndpoint  string    `json:"minio_endpoint"`
	MinioPublicURL string    `json:"minio_public_url"`
	MinioAccessKey string    `json:"minio_access_key"`
	MinioSecretKey string    `json:"minio_secret_key"`
	MinioUseSSL    bool      `json:"minio_use_ssl"`
	MergeMethod    string    `json:"merge_method"`
	FastMode       bool      `json:"fast_mode"`
	GenerationMode string    `json:"generation_mode"`
	CreatedAt      time.Time `json:"created_at"`
	Priority       int       `json:"priority"`
}
