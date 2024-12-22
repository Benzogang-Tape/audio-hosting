package model

import "github.com/google/uuid"

type UserUser struct {
	FollowerID uuid.UUID `json:"follower_id"`
	FollowedID uuid.UUID `json:"followed_id"`
}
