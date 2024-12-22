package users

import "github.com/google/uuid"

type Artist struct {
	Id        uuid.UUID
	Name      string
	Label     string
	AvatarUrl string
}
