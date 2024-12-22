package listeners

import "context"

type DTOGetListenerInput struct{}

type DTOGetListenerOutput struct{}

func (se *Service) GetListener(
	_ context.Context,
	_ DTOGetListenerInput,
) (DTOGetListenerOutput, error) {
	panic("not implemented") // TODO: Implement
}
