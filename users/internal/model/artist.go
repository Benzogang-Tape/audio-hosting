package model

type Artist struct {
	User
	Label string `json:"label"`
}
