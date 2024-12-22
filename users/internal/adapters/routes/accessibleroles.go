package routes

type AdapterType string

const (
	GRPC AdapterType = "grpc"
	HTTP AdapterType = "http"

	grpcPath = "/users_api.UsersService/"
	httpPath = "/users/api/v1/"
)

func PublicRoutes(adapterType AdapterType) []string {
	switch adapterType {
	case GRPC:
		return []string{
			grpcPath + "Register",
			grpcPath + "Refresh",
			grpcPath + "Login",
			grpcPath + "Logout",
			grpcPath + "GetUser",
			grpcPath + "GetArtists",
		}
	case HTTP:
		return []string{
			httpPath + "users/login",
			httpPath + "users/register",
			httpPath + "users/refresh",
			httpPath + "users/logout",
		}
	default:
		return []string{}
	}
}
