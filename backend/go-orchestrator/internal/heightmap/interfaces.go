package heightmap

import (
	"context"
	"io"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/skr1ms/dev2gis/internal/storage/sqlc"
)

type MinioClientInterface interface {
	UploadFile(ctx context.Context, bucket, objectName string, reader io.Reader, size int64, contentType string) error
	FileExists(ctx context.Context, bucket, objectName string) (bool, error)
	GetPresignedURL(ctx context.Context, bucket, objectName string, expiry int) (string, error)
}

type QueriesInterface interface {
	CreateHeightmapJob(ctx context.Context, params sqlc.CreateHeightmapJobParams) (sqlc.HeightmapJob, error)
	GetHeightmapJob(ctx context.Context, id uuid.UUID) (sqlc.HeightmapJob, error)
	GetHeightmapJobByUserID(ctx context.Context, params sqlc.GetHeightmapJobByUserIDParams) (sqlc.HeightmapJob, error)
	ListUserHeightmaps(ctx context.Context, params sqlc.ListUserHeightmapsParams) ([]sqlc.HeightmapJob, error)
	UpdateJobStatus(ctx context.Context, params sqlc.UpdateJobStatusParams) error
	UpdateJobResult(ctx context.Context, params sqlc.UpdateJobResultParams) error
	UpdateJobError(ctx context.Context, params sqlc.UpdateJobErrorParams) error

	CreateBatchHeightmapJob(ctx context.Context, params sqlc.CreateBatchHeightmapJobParams) (sqlc.BatchHeightmapJob, error)
	CreateBatchImage(ctx context.Context, params sqlc.CreateBatchImageParams) (sqlc.BatchImage, error)
	GetBatchHeightmapJob(ctx context.Context, id uuid.UUID) (sqlc.BatchHeightmapJob, error)
	GetBatchHeightmapJobByUserID(ctx context.Context, params sqlc.GetBatchHeightmapJobByUserIDParams) (sqlc.BatchHeightmapJob, error)
	ListUserBatchHeightmaps(ctx context.Context, params sqlc.ListUserBatchHeightmapsParams) ([]sqlc.BatchHeightmapJob, error)
	UpdateBatchJobStatus(ctx context.Context, params sqlc.UpdateBatchJobStatusParams) error
	UpdateBatchJobResult(ctx context.Context, params sqlc.UpdateBatchJobResultParams) error
	UpdateBatchJobError(ctx context.Context, params sqlc.UpdateBatchJobErrorParams) error
	UpdateBatchJobProgress(ctx context.Context, params sqlc.UpdateBatchJobProgressParams) error
	GetBatchImages(ctx context.Context, batchJobID pgtype.UUID) ([]sqlc.BatchImage, error)
	UpdateBatchImageStatus(ctx context.Context, params sqlc.UpdateBatchImageStatusParams) error
}
