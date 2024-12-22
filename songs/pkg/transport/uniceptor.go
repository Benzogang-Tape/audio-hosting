package transport

import "context"

type GrpcHandler[T any, T2 any] func(context.Context, T) (T2, error)
type Uniceptor[T any, T2 any] func(next GrpcHandler[T, T2]) GrpcHandler[T, T2]
type HandInvoker[T any, T2 any] func(GrpcHandler[T, T2]) (T2, error)

func Apply[T any, T2 any](ctx context.Context, req T, unis ...Uniceptor[T, T2]) HandInvoker[T, T2] {
	return func(handler GrpcHandler[T, T2]) (T2, error) {
		for i := len(unis) - 1; i >= 0; i-- {
			handler = unis[i](handler)
		}

		return handler(ctx, req)
	}
}
