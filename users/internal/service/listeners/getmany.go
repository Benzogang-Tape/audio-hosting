package listeners

import "context"

type DTOGetListenersInput struct{}

type DTOGetListenersOutput struct{}

func (se *Service) GetListeners(
	_ context.Context,
	_ DTOGetListenersInput,
) (DTOGetListenersOutput, error) {
	panic("not implemented") // TODO: Implement
}
