package users

import "context"

type DTOGetFollowersInput struct{}

type DTOGetFollowersOutput struct{}

func (*Service) GetFollowers(
	_ context.Context,
	_ DTOGetFollowersInput,
) (DTOGetFollowersOutput, error) {
	panic("not implemented") // TODO: Implement
}
