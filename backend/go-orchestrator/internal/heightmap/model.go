package heightmap

import (
	"time"

	"github.com/google/uuid"
)

type HeightmapJob struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	ImageURL       string    `json:"image_url"`
	ResultURL      *string   `json:"result_url,omitempty"`
	Status         string    `json:"status"`
	Width          *int32    `json:"width,omitempty"`
	Height         *int32    `json:"height,omitempty"`
	ErrorMessage   *string   `json:"error_message,omitempty"`
	ProcessingTime *float64  `json:"processing_time,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type PointHeight struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Height float64 `json:"height"`
}
