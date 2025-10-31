package heightmap

import "github.com/google/uuid"

type UploadRequest struct {
	Name string `json:"name" form:"name" binding:"required"`
}

type UploadResponse struct {
	ID     uuid.UUID `json:"id"`
	Status string    `json:"status"`
}

type HeightRequest struct {
	X float64 `json:"x" form:"x" binding:"required"`
	Y float64 `json:"y" form:"y" binding:"required"`
}

type HeightResponse struct {
	Height float64 `json:"height"`
}

type ProfileRequest struct {
	StartPoint Point `json:"start_point"`
	EndPoint   Point `json:"end_point"`
}

type ProfileResponse struct {
	Points []PointHeight `json:"points"`
}

type BatchUploadResponse struct {
	ID          uuid.UUID `json:"id"`
	Status      string    `json:"status"`
	ImageCount  int       `json:"image_count"`
	MergeMethod string    `json:"merge_method"`
}

type BatchHeightmapJob struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	Status         string    `json:"status"`
	ResultURL      *string   `json:"result_url,omitempty"`
	OrthophotoURL  *string   `json:"orthophoto_url,omitempty"`
	Width          *int32    `json:"width,omitempty"`
	Height         *int32    `json:"height,omitempty"`
	ImageCount     int32     `json:"image_count"`
	ProcessedCount int32     `json:"processed_count"`
	ErrorMessage   *string   `json:"error_message,omitempty"`
	ProcessingTime *float64  `json:"processing_time,omitempty"`
	MergeMethod    string    `json:"merge_method"`
	CreatedAt      string    `json:"created_at"`
	UpdatedAt      string    `json:"updated_at"`
}
