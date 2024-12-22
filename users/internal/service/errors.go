package service

import "errors"

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrWrongPassword     = errors.New("wrong password")
	ErrUserAlreadyExists = errors.New("user already exists")

	ErrSessionExpired = errors.New("session expired")

	ErrListenerNotFound = errors.New("listener not found")
	ErrArtistNotFound   = errors.New("artist not found")
)
