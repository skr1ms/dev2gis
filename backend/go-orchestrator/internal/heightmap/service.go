package heightmap

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/skr1ms/dev2gis/config"
	"github.com/skr1ms/dev2gis/internal/storage"
	"github.com/skr1ms/dev2gis/internal/storage/minio"
	"github.com/skr1ms/dev2gis/internal/storage/sqlc"
	"github.com/skr1ms/dev2gis/pkg/metrics"
	"github.com/skr1ms/dev2gis/pkg/rabbitmq"
)

type Service struct {
	db             *storage.DB
	queries        QueriesInterface
	minioClient    MinioClientInterface
	rabbitmqClient *rabbitmq.Client
	cfg            *config.Config
}

func NewService(db *storage.DB, minioClient *minio.MinioClient, rabbitmqClient *rabbitmq.Client, cfg *config.Config) *Service {
	return &Service{
		db:             db,
		queries:        db.Queries,
		minioClient:    minioClient,
		rabbitmqClient: rabbitmqClient,
		cfg:            cfg,
	}
}

func (s *Service) UploadPhoto(ctx context.Context, userID uuid.UUID, file multipart.File, header *multipart.FileHeader) (*UploadResponse, error) {
	start := time.Now()
	metrics.RecordProcessingJob()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return nil, fmt.Errorf("неподдерживаемый формат файла: %s", ext)
	}

	if header.Size > 100*1024*1024 {
		return nil, fmt.Errorf("файл слишком большой: максимум 100МБ")
	}

	jobID := uuid.New()
	objectName := fmt.Sprintf("heightmaps/%s/%s%s", userID.String(), jobID.String(), ext)

	if err := s.minioClient.UploadFile(ctx, s.cfg.Minio.UAVDataBucketName, objectName, file, header.Size, header.Header.Get("Content-Type")); err != nil {
		return nil, fmt.Errorf("не удалось загрузить файл в хранилище: %w", err)
	}

	imageURL := fmt.Sprintf("%s/%s/%s", s.cfg.Minio.PublicURL, s.cfg.Minio.UAVDataBucketName, objectName)

	now := time.Now()
	job := sqlc.CreateHeightmapJobParams{
		ID:        jobID,
		UserID:    userID,
		ImageUrl:  imageURL,
		Status:    "pending",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if _, err := s.queries.CreateHeightmapJob(ctx, job); err != nil {
		return nil, fmt.Errorf("не удалось создать задачу в базе данных: %w", err)
	}

	task := &rabbitmq.HeightmapTask{
		JobID:          jobID.String(),
		UserID:         userID.String(),
		ImageURL:       imageURL,
		OutputBucket:   s.cfg.Minio.UAVModelsBucketName,
		MinioEndpoint:  s.cfg.Minio.Endpoint,
		MinioPublicURL: s.cfg.Minio.PublicURL,
		MinioAccessKey: s.cfg.Minio.AccessKeyID,
		MinioSecretKey: s.cfg.Minio.SecretAccessKey,
		MinioUseSSL:    s.cfg.Minio.UseSSL,
		CreatedAt:      now,
		Priority:       0,
	}

	if err := s.rabbitmqClient.PublishTask(ctx, task); err != nil {
		now := time.Now()
		errMsg := fmt.Sprintf("не удалось отправить задачу в очередь: %v", err)
		_ = s.queries.UpdateJobError(ctx, sqlc.UpdateJobErrorParams{
			ID:           jobID,
			ErrorMessage: &errMsg,
			UpdatedAt:    now,
		})
		return nil, fmt.Errorf("не удалось отправить задачу: %w", err)
	}

	metrics.RecordProcessingJobDuration(time.Since(start))

	return &UploadResponse{
		ID:     jobID,
		Status: "pending",
	}, nil
}

func (s *Service) GetHeightmapJob(ctx context.Context, jobID uuid.UUID, userID uuid.UUID) (*HeightmapJob, error) {
	job, err := s.queries.GetHeightmapJobByUserID(ctx, sqlc.GetHeightmapJobByUserIDParams{
		ID:     jobID,
		UserID: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("карта высот не найдена: %w", err)
	}

	result := &HeightmapJob{
		ID:             job.ID,
		UserID:         job.UserID,
		ImageURL:       job.ImageUrl,
		ResultURL:      job.ResultUrl,
		Status:         job.Status,
		Width:          job.Width,
		Height:         job.Height,
		ErrorMessage:   job.ErrorMessage,
		ProcessingTime: job.ProcessingTime,
		CreatedAt:      job.CreatedAt,
		UpdatedAt:      job.UpdatedAt,
	}

	return result, nil
}

func (s *Service) ListUserHeightmaps(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]*HeightmapJob, error) {
	jobs, err := s.queries.ListUserHeightmaps(ctx, sqlc.ListUserHeightmapsParams{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("не удалось получить список карт высот: %w", err)
	}

	result := make([]*HeightmapJob, 0, len(jobs))
	for _, job := range jobs {
		hm := &HeightmapJob{
			ID:             job.ID,
			UserID:         job.UserID,
			ImageURL:       job.ImageUrl,
			ResultURL:      job.ResultUrl,
			Status:         job.Status,
			Width:          job.Width,
			Height:         job.Height,
			ErrorMessage:   job.ErrorMessage,
			ProcessingTime: job.ProcessingTime,
			CreatedAt:      job.CreatedAt,
			UpdatedAt:      job.UpdatedAt,
		}

		result = append(result, hm)
	}

	return result, nil
}

func (s *Service) GetHeight(ctx context.Context, id uuid.UUID, x, y float64) (*HeightResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *Service) GetProfile(ctx context.Context, id uuid.UUID, startPoint, endPoint Point) (*ProfileResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *Service) BatchUploadPhotos(ctx context.Context, userID uuid.UUID, files []*multipart.FileHeader, mergeMethod string, fastMode bool, generationMode string) (*BatchUploadResponse, error) {
	start := time.Now()
	metrics.RecordProcessingJob()

	if len(files) == 0 {
		return nil, fmt.Errorf("файлы не предоставлены")
	}

	if len(files) > 50 {
		return nil, fmt.Errorf("слишком много файлов: максимум 50 файлов в пакете")
	}

	if mergeMethod == "" {
		mergeMethod = "medium"
	}
	if mergeMethod != "medium" && mergeMethod != "max" && mergeMethod != "low" {
		return nil, fmt.Errorf("некорректный метод объединения: %s (разрешены: low, medium, max)", mergeMethod)
	}

	if generationMode == "" {
		generationMode = "heightmap"
	}
	if generationMode != "heightmap" && generationMode != "orthophoto" && generationMode != "both" {
		return nil, fmt.Errorf("некорректный режим генерации: %s (разрешены: heightmap, orthophoto, both)", generationMode)
	}

	if (generationMode == "orthophoto" || generationMode == "both") && len(files) < 5 {
		return nil, fmt.Errorf("для генерации ортофотоплана требуется минимум 5 изображений, предоставлено: %d", len(files))
	}

	batchJobID := uuid.New()
	now := time.Now()

	batchJob := sqlc.CreateBatchHeightmapJobParams{
		ID:             batchJobID,
		UserID:         userID,
		Status:         "pending",
		ImageCount:     int32(len(files)),
		MergeMethod:    mergeMethod,
		GenerationMode: generationMode,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if _, err := s.queries.CreateBatchHeightmapJob(ctx, batchJob); err != nil {
		return nil, fmt.Errorf("не удалось создать пакетную задачу в базе данных: %w", err)
	}

	imageURLs := make([]string, 0, len(files))

	for idx, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			return nil, fmt.Errorf("не удалось открыть файл %s: %w", fileHeader.Filename, err)
		}

		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
			file.Close()
			return nil, fmt.Errorf("неподдерживаемый формат файла %s: %s", fileHeader.Filename, ext)
		}

		if fileHeader.Size > 100*1024*1024 {
			file.Close()
			return nil, fmt.Errorf("файл %s слишком большой: максимум 100МБ", fileHeader.Filename)
		}

		imageID := uuid.New()
		objectName := fmt.Sprintf("batch-heightmaps/%s/%s/image_%d%s", userID.String(), batchJobID.String(), idx, ext)

		if err := s.minioClient.UploadFile(ctx, s.cfg.Minio.UAVDataBucketName, objectName, file, fileHeader.Size, fileHeader.Header.Get("Content-Type")); err != nil {
			file.Close()
			return nil, fmt.Errorf("не удалось загрузить файл %s в хранилище: %w", fileHeader.Filename, err)
		}
		file.Close()

		imageURL := fmt.Sprintf("%s/%s/%s", s.cfg.Minio.PublicURL, s.cfg.Minio.UAVDataBucketName, objectName)
		imageURLs = append(imageURLs, imageURL)

		batchImage := sqlc.CreateBatchImageParams{
			ID:        imageID,
			ImageUrl:  imageURL,
			Status:    "pending",
			CreatedAt: now,
		}
		batchImage.BatchJobID.Bytes = batchJobID
		batchImage.BatchJobID.Valid = true

		if _, err := s.queries.CreateBatchImage(ctx, batchImage); err != nil {
			return nil, fmt.Errorf("не удалось создать запись изображения в базе данных: %w", err)
		}
	}

	task := &rabbitmq.BatchHeightmapTask{
		BatchJobID:     batchJobID.String(),
		UserID:         userID.String(),
		ImageURLs:      imageURLs,
		OutputBucket:   s.cfg.Minio.UAVModelsBucketName,
		MinioEndpoint:  s.cfg.Minio.Endpoint,
		MinioPublicURL: s.cfg.Minio.PublicURL,
		MinioAccessKey: s.cfg.Minio.AccessKeyID,
		MinioSecretKey: s.cfg.Minio.SecretAccessKey,
		MinioUseSSL:    s.cfg.Minio.UseSSL,
		MergeMethod:    mergeMethod,
		FastMode:       fastMode,
		GenerationMode: generationMode,
		CreatedAt:      now,
		Priority:       0,
	}

	if err := s.rabbitmqClient.PublishBatchTask(ctx, task); err != nil {
		now := time.Now()
		errMsg := fmt.Sprintf("не удалось отправить пакетную задачу в очередь: %v", err)
		_ = s.queries.UpdateBatchJobError(ctx, sqlc.UpdateBatchJobErrorParams{
			ID:           batchJobID,
			ErrorMessage: &errMsg,
			UpdatedAt:    now,
		})
		return nil, fmt.Errorf("не удалось отправить пакетную задачу: %w", err)
	}

	metrics.RecordProcessingJobDuration(time.Since(start))

	return &BatchUploadResponse{
		ID:          batchJobID,
		Status:      "pending",
		ImageCount:  len(files),
		MergeMethod: mergeMethod,
	}, nil
}

func (s *Service) GetBatchHeightmapJob(ctx context.Context, batchJobID uuid.UUID, userID uuid.UUID) (*BatchHeightmapJob, error) {
	job, err := s.queries.GetBatchHeightmapJobByUserID(ctx, sqlc.GetBatchHeightmapJobByUserIDParams{
		ID:     batchJobID,
		UserID: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("пакетная карта высот не найдена: %w", err)
	}

	result := &BatchHeightmapJob{
		ID:             job.ID,
		UserID:         job.UserID,
		Status:         job.Status,
		ResultURL:      job.ResultUrl,
		OrthophotoURL:  job.OrthophotoUrl,
		Width:          job.Width,
		Height:         job.Height,
		ImageCount:     job.ImageCount,
		ProcessedCount: job.ProcessedCount,
		ErrorMessage:   job.ErrorMessage,
		ProcessingTime: job.ProcessingTime,
		MergeMethod:    job.MergeMethod,
		CreatedAt:      job.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      job.UpdatedAt.Format(time.RFC3339),
	}

	return result, nil
}

func (s *Service) ListUserBatchHeightmaps(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]*BatchHeightmapJob, error) {
	jobs, err := s.queries.ListUserBatchHeightmaps(ctx, sqlc.ListUserBatchHeightmapsParams{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("не удалось получить список пакетных карт высот: %w", err)
	}

	result := make([]*BatchHeightmapJob, 0, len(jobs))
	for _, job := range jobs {
		hm := &BatchHeightmapJob{
			ID:             job.ID,
			UserID:         job.UserID,
			Status:         job.Status,
			ResultURL:      job.ResultUrl,
			OrthophotoURL:  job.OrthophotoUrl,
			Width:          job.Width,
			Height:         job.Height,
			ImageCount:     job.ImageCount,
			ProcessedCount: job.ProcessedCount,
			ErrorMessage:   job.ErrorMessage,
			ProcessingTime: job.ProcessingTime,
			MergeMethod:    job.MergeMethod,
			CreatedAt:      job.CreatedAt.Format(time.RFC3339),
			UpdatedAt:      job.UpdatedAt.Format(time.RFC3339),
		}

		result = append(result, hm)
	}

	return result, nil
}
