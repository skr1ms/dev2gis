package minio

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/skr1ms/dev2gis/config"
	"github.com/skr1ms/dev2gis/pkg/middleware"
)

type MinioClient = Minio

type Minio struct {
	client          *minio.Client
	uavDataBucket   string
	uavModelsBucket string
	publicURL       string
	logger          middleware.LoggerInterface
}

type FileInfo struct {
	Name         string
	Size         int64
	ContentType  string
	LastModified time.Time
	ETag         string
}

func NewMinio(cfg config.MinioConfig, logger middleware.LoggerInterface) (*Minio, error) {
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	minio := &Minio{
		client:          minioClient,
		uavDataBucket:   cfg.UAVDataBucketName,
		uavModelsBucket: cfg.UAVModelsBucketName,
		publicURL:       cfg.PublicURL,
		logger:          logger,
	}

	if err := minio.ensureBucketExists(context.Background(), minio.uavDataBucket); err != nil {
		return nil, fmt.Errorf("failed to ensure bucket exists: %w", err)
	}

	if minio.uavModelsBucket != "" || minio.uavDataBucket != "" {
		if err := minio.ensureBucketExists(context.Background(), minio.uavModelsBucket); err != nil {
			return nil, fmt.Errorf("failed to ensure uav models bucket exists: %w", err)
		}
	}

	return minio, nil
}

func (m *Minio) ensureBucketExists(ctx context.Context, bucket string) error {
	exists, err := m.client.BucketExists(ctx, bucket)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = m.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
		m.logger.Info("Created MinIO bucket", map[string]interface{}{
			"bucket": bucket,
		})
	}

	m.logger.Info("MinIO bucket exists", map[string]interface{}{
		"bucket": bucket,
	})
	return nil
}

func (m *Minio) UploadFile(ctx context.Context, bucket, objectName string, reader io.Reader, objectSize int64, contentType string) error {
	_, err := m.client.PutObject(ctx, bucket, objectName, reader, objectSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	m.logger.Info("File uploaded to MinIO", map[string]interface{}{
		"bucket": bucket,
		"object": objectName,
		"size":   objectSize,
	})
	return nil
}

func (m *Minio) GetFileURL(bucket, objectName string) string {
	return fmt.Sprintf("%s/%s/%s", m.publicURL, bucket, objectName)
}

func (m *Minio) GetPresignedURL(ctx context.Context, bucket, objectName string, expirySeconds int) (string, error) {
	expiry := time.Duration(expirySeconds) * time.Second
	url, err := m.client.PresignedGetObject(ctx, bucket, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	m.logger.Info("Generated presigned URL", map[string]interface{}{
		"bucket":         bucket,
		"object":         objectName,
		"expiry_seconds": expirySeconds,
	})

	return url.String(), nil
}

func (m *Minio) FileExists(ctx context.Context, bucket, objectName string) (bool, error) {
	_, err := m.client.StatObject(ctx, bucket, objectName, minio.StatObjectOptions{})
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "NoSuchKey" {
			return false, nil
		}
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}
	return true, nil
}

func (m *Minio) GetFileInfo(ctx context.Context, bucket, objectName string) (*FileInfo, error) {
	stat, err := m.client.StatObject(ctx, bucket, objectName, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return &FileInfo{
		Name:         objectName,
		Size:         stat.Size,
		ContentType:  stat.ContentType,
		LastModified: stat.LastModified,
		ETag:         stat.ETag,
	}, nil
}
