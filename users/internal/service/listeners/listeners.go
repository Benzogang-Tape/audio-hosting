package listeners

type ListenersRepository interface{}

type Service struct {
	listenersRepository ListenersRepository
}

func New(listenersRepository ListenersRepository) *Service {
	return &Service{
		listenersRepository: listenersRepository,
	}
}
