package listeners

import "context"

type DTODeleteListenerInput struct{}

func (se *Service) DeleteListener(_ context.Context, _ DTODeleteListenerInput) error {
	panic("not implemented") // TODO: Implement
}
