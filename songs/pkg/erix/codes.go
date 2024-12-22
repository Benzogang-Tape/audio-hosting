package erix

import (
	"google.golang.org/grpc/codes"
)

const (
	CodeBadRequest = iota
	CodeUnauthorized
	CodeForbidden
	CodeNotFound
	CodeRequestTimeout
	CodeConflict
	CodePreconditionFailed
	CodeTooManyRequests
	CodeInternalServerError
	CodeNotImplemented
	CodeServiceUnavailable
	CodeGatewayTimeout
)

var (
	httpMap = []int{400, 401, 403, 404, 408, 409, 412, 429, 500, 501, 503, 504}
	grpcMap = []codes.Code{
		codes.InvalidArgument,
		codes.Unauthenticated,
		codes.PermissionDenied,
		codes.NotFound,
		codes.Canceled,
		codes.AlreadyExists,
		codes.FailedPrecondition,
		codes.ResourceExhausted,
		codes.Internal,
		codes.Unimplemented,
		codes.Unavailable,
		codes.DeadlineExceeded,
	}
)

func ToHttp(code int) int {
	return httpMap[code]
}

func ToGrpc(code int) codes.Code {
	return grpcMap[code]
}
