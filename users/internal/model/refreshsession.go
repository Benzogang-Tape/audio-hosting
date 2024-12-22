package model

import (
	"time"

	"github.com/google/uuid"
)

type RefreshSession struct {
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	UserID       uuid.UUID `json:"user_id"`
}
