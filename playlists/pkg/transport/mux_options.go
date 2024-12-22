package transport

import gateway "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

func MuxWithAuthAndTraceHeaders() gateway.ServeMuxOption {
	return gateway.WithIncomingHeaderMatcher(func(key string) (string, bool) {
		switch key {
		case AuthKey:
			return key, true

		case TraceIdKey:
			return key, true
		}

		return key, false
	})
}
