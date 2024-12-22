package model

import "github.com/google/uuid"

type NotificationSettings struct {
	UserID             uuid.UUID `json:"user_id"`
	EmailNotifications bool      `json:"email_notifications"`
}
