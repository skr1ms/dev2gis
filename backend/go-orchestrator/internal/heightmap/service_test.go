package heightmap

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/skr1ms/dev2gis/config"
	"github.com/skr1ms/dev2gis/internal/storage/sqlc"
)

type mockQueries struct {
	createJobFunc       func(ctx context.Context, params sqlc.CreateHeightmapJobParams) (sqlc.HeightmapJob, error)
	getJobFunc          func(ctx context.Context, id uuid.UUID) (sqlc.HeightmapJob, error)
	getJobByUserIDFunc  func(ctx context.Context, params sqlc.GetHeightmapJobByUserIDParams) (sqlc.HeightmapJob, error)
	listUserHeightmaps  func(ctx context.Context, params sqlc.ListUserHeightmapsParams) ([]sqlc.HeightmapJob, error)
	updateJobStatusFunc func(ctx context.Context, params sqlc.UpdateJobStatusParams) error
	updateJobResultFunc func(ctx context.Context, params sqlc.UpdateJobResultParams) error
	updateJobErrorFunc  func(ctx context.Context, params sqlc.UpdateJobErrorParams) error
}

func (m *mockQueries) CreateHeightmapJob(ctx context.Context, params sqlc.CreateHeightmapJobParams) (sqlc.HeightmapJob, error) {
	if m.createJobFunc != nil {
		return m.createJobFunc(ctx, params)
	}
	return sqlc.HeightmapJob{}, nil
}

func (m *mockQueries) GetHeightmapJob(ctx context.Context, id uuid.UUID) (sqlc.HeightmapJob, error) {
	if m.getJobFunc != nil {
		return m.getJobFunc(ctx, id)
	}
	return sqlc.HeightmapJob{}, nil
}

func (m *mockQueries) GetHeightmapJobByUserID(ctx context.Context, params sqlc.GetHeightmapJobByUserIDParams) (sqlc.HeightmapJob, error) {
	if m.getJobByUserIDFunc != nil {
		return m.getJobByUserIDFunc(ctx, params)
	}
	return sqlc.HeightmapJob{}, nil
}

func (m *mockQueries) ListUserHeightmaps(ctx context.Context, params sqlc.ListUserHeightmapsParams) ([]sqlc.HeightmapJob, error) {
	if m.listUserHeightmaps != nil {
		return m.listUserHeightmaps(ctx, params)
	}
	return []sqlc.HeightmapJob{}, nil
}

func (m *mockQueries) UpdateJobStatus(ctx context.Context, params sqlc.UpdateJobStatusParams) error {
	if m.updateJobStatusFunc != nil {
		return m.updateJobStatusFunc(ctx, params)
	}
	return nil
}

func (m *mockQueries) UpdateJobResult(ctx context.Context, params sqlc.UpdateJobResultParams) error {
	if m.updateJobResultFunc != nil {
		return m.updateJobResultFunc(ctx, params)
	}
	return nil
}

func (m *mockQueries) UpdateJobError(ctx context.Context, params sqlc.UpdateJobErrorParams) error {
	if m.updateJobErrorFunc != nil {
		return m.updateJobErrorFunc(ctx, params)
	}
	return nil
}

func (m *mockQueries) CreateBatchHeightmapJob(ctx context.Context, params sqlc.CreateBatchHeightmapJobParams) (sqlc.BatchHeightmapJob, error) {
	return sqlc.BatchHeightmapJob{}, nil
}

func (m *mockQueries) CreateBatchImage(ctx context.Context, params sqlc.CreateBatchImageParams) (sqlc.BatchImage, error) {
	return sqlc.BatchImage{}, nil
}

func (m *mockQueries) GetBatchHeightmapJob(ctx context.Context, id uuid.UUID) (sqlc.BatchHeightmapJob, error) {
	return sqlc.BatchHeightmapJob{}, nil
}

func (m *mockQueries) GetBatchHeightmapJobByUserID(ctx context.Context, params sqlc.GetBatchHeightmapJobByUserIDParams) (sqlc.BatchHeightmapJob, error) {
	return sqlc.BatchHeightmapJob{}, nil
}

func (m *mockQueries) ListUserBatchHeightmaps(ctx context.Context, params sqlc.ListUserBatchHeightmapsParams) ([]sqlc.BatchHeightmapJob, error) {
	return []sqlc.BatchHeightmapJob{}, nil
}

func (m *mockQueries) UpdateBatchJobStatus(ctx context.Context, params sqlc.UpdateBatchJobStatusParams) error {
	return nil
}

func (m *mockQueries) UpdateBatchJobResult(ctx context.Context, params sqlc.UpdateBatchJobResultParams) error {
	return nil
}

func (m *mockQueries) UpdateBatchJobError(ctx context.Context, params sqlc.UpdateBatchJobErrorParams) error {
	return nil
}

func (m *mockQueries) UpdateBatchJobProgress(ctx context.Context, params sqlc.UpdateBatchJobProgressParams) error {
	return nil
}

func (m *mockQueries) GetBatchImages(ctx context.Context, batchJobID pgtype.UUID) ([]sqlc.BatchImage, error) {
	return []sqlc.BatchImage{}, nil
}

func (m *mockQueries) UpdateBatchImageStatus(ctx context.Context, params sqlc.UpdateBatchImageStatusParams) error {
	return nil
}

type mockMinioClient struct {
	uploadFileFunc      func(ctx context.Context, bucket, objectName string, reader io.Reader, size int64, contentType string) error
	fileExistsFunc      func(ctx context.Context, bucket, objectName string) (bool, error)
	getPresignedURLFunc func(ctx context.Context, bucket, objectName string, expiry int) (string, error)
}

func (m *mockMinioClient) UploadFile(ctx context.Context, bucket, objectName string, reader io.Reader, size int64, contentType string) error {
	if m.uploadFileFunc != nil {
		return m.uploadFileFunc(ctx, bucket, objectName, reader, size, contentType)
	}
	return nil
}

func (m *mockMinioClient) FileExists(ctx context.Context, bucket, objectName string) (bool, error) {
	if m.fileExistsFunc != nil {
		return m.fileExistsFunc(ctx, bucket, objectName)
	}
	return true, nil
}

func (m *mockMinioClient) GetPresignedURL(ctx context.Context, bucket, objectName string, expiry int) (string, error) {
	if m.getPresignedURLFunc != nil {
		return m.getPresignedURLFunc(ctx, bucket, objectName, expiry)
	}
	return "https://minio.example.com/presigned-url", nil
}

func TestGetHeightmapJob(t *testing.T) {
	jobID := uuid.New()
	userID := uuid.New()
	resultURL := "http://minio:9000/uav-models/heightmaps/test.png"

	tests := []struct {
		name        string
		jobStatus   string
		hasResult   bool
		expectError bool
	}{
		{
			name:        "completed job with result",
			jobStatus:   "completed",
			hasResult:   true,
			expectError: false,
		},
		{
			name:        "pending job",
			jobStatus:   "pending",
			hasResult:   false,
			expectError: false,
		},
		{
			name:        "job not found",
			jobStatus:   "",
			hasResult:   false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queries := &mockQueries{
				getJobByUserIDFunc: func(ctx context.Context, params sqlc.GetHeightmapJobByUserIDParams) (sqlc.HeightmapJob, error) {
					if tt.expectError {
						return sqlc.HeightmapJob{}, errors.New("not found")
					}

					job := sqlc.HeightmapJob{
						ID:        jobID,
						UserID:    userID,
						Status:    tt.jobStatus,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					}

					if tt.hasResult {
						job.ResultUrl = &resultURL
					}

					return job, nil
				},
			}

			mockMinio := &mockMinioClient{
				fileExistsFunc: func(ctx context.Context, bucket, objectName string) (bool, error) {
					return true, nil
				},
				getPresignedURLFunc: func(ctx context.Context, bucket, objectName string, expiry int) (string, error) {
					return "https://presigned.url", nil
				},
			}

			cfg := &config.Config{
				Minio: config.MinioConfig{
					UAVModelsBucketName: "uav-models",
				},
			}

			s := &Service{
				queries:     queries,
				minioClient: mockMinio,
				cfg:         cfg,
			}

			result, err := s.GetHeightmapJob(context.Background(), jobID, userID)

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.expectError && result == nil {
				t.Errorf("expected result but got nil")
			}

			if !tt.expectError && result != nil {
				if result.Status != tt.jobStatus {
					t.Errorf("expected status %s, got %s", tt.jobStatus, result.Status)
				}

				if tt.hasResult && tt.jobStatus == "completed" {
					if result.ResultURL == nil {
						t.Errorf("expected result URL but got nil")
					}
				}
			}
		})
	}
}

func TestListUserHeightmaps(t *testing.T) {
	userID := uuid.New()
	resultURL := "http://minio:9000/uav-models/heightmaps/test.png"

	tests := []struct {
		name        string
		jobCount    int
		expectError bool
	}{
		{
			name:        "list with results",
			jobCount:    3,
			expectError: false,
		},
		{
			name:        "empty list",
			jobCount:    0,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queries := &mockQueries{
				listUserHeightmaps: func(ctx context.Context, params sqlc.ListUserHeightmapsParams) ([]sqlc.HeightmapJob, error) {
					if tt.expectError {
						return nil, errors.New("database error")
					}

					jobs := make([]sqlc.HeightmapJob, tt.jobCount)
					for i := 0; i < tt.jobCount; i++ {
						jobs[i] = sqlc.HeightmapJob{
							ID:        uuid.New(),
							UserID:    userID,
							Status:    "completed",
							ResultUrl: &resultURL,
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						}
					}

					return jobs, nil
				},
			}

			mockMinio := &mockMinioClient{
				fileExistsFunc: func(ctx context.Context, bucket, objectName string) (bool, error) {
					return true, nil
				},
				getPresignedURLFunc: func(ctx context.Context, bucket, objectName string, expiry int) (string, error) {
					return "https://presigned.url", nil
				},
			}

			cfg := &config.Config{
				Minio: config.MinioConfig{
					UAVModelsBucketName: "uav-models",
				},
			}

			s := &Service{
				queries:     queries,
				minioClient: mockMinio,
				cfg:         cfg,
			}

			results, err := s.ListUserHeightmaps(context.Background(), userID, 20, 0)

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.expectError && len(results) != tt.jobCount {
				t.Errorf("expected %d results, got %d", tt.jobCount, len(results))
			}
		})
	}
}

func TestBatchHeightmapJobMethods(t *testing.T) {
	t.Run("GetBatchHeightmapJob should not panic", func(t *testing.T) {
		batchJobID := uuid.New()
		userID := uuid.New()

		queries := &mockQueries{}
		mockMinio := &mockMinioClient{}
		cfg := &config.Config{
			Minio: config.MinioConfig{
				UAVModelsBucketName: "uav-models",
			},
		}

		s := &Service{
			queries:     queries,
			minioClient: mockMinio,
			cfg:         cfg,
		}

		_, err := s.GetBatchHeightmapJob(context.Background(), batchJobID, userID)
		if err == nil {
			t.Log("GetBatchHeightmapJob executed without panic")
		}
	})

	t.Run("ListUserBatchHeightmaps should not panic", func(t *testing.T) {
		userID := uuid.New()

		queries := &mockQueries{}
		mockMinio := &mockMinioClient{}
		cfg := &config.Config{
			Minio: config.MinioConfig{
				UAVModelsBucketName: "uav-models",
			},
		}

		s := &Service{
			queries:     queries,
			minioClient: mockMinio,
			cfg:         cfg,
		}

		_, err := s.ListUserBatchHeightmaps(context.Background(), userID, 20, 0)
		if err == nil {
			t.Log("ListUserBatchHeightmaps executed without panic")
		}
	})
}
