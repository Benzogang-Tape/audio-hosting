package users

import "context"

type DTOGetFollowedInput struct{}

type DTOGetFollowedOutput struct{}

func (*Service) GetFollowed(
	_ context.Context,
	_ DTOGetFollowedInput,
) (DTOGetFollowedOutput, error) {
	panic("not implemented") // TODO: Implement
}
