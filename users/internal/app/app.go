package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"github.com/Benzogang-Tape/audio-hosting/users/api/protogen"
	"github.com/Benzogang-Tape/audio-hosting/users/internal/adapters/grpc/interceptors"
	"github.com/Benzogang-Tape/audio-hosting/users/internal/adapters/http/middlewares"
	"github.com/Benzogang-Tape/audio-hosting/users/internal/adapters/routes"
	"github.com/Benzogang-Tape/audio-hosting/users/migrations"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

type App struct {
	provider *Provider

	grpcServer *grpc.Server
	httpServer *http.Server
}

func NewApp(ctx context.Context) (*App, error) {
	a := &App{}

	err := a.initDeps(ctx)
	if err != nil {
		return nil, fmt.Errorf("app.NewApp: %w", err)
	}

	return a, nil
}

func (a *App) Run(ctx context.Context) error {
	var eg errgroup.Group

	eg.Go(func() error {
		return a.runGRPCServer(ctx)
	})

	eg.Go(func() error {
		return a.runHTTPServer(ctx)
	})

	return eg.Wait()
}

func (a *App) GracefulShutdown(ctx context.Context) error {
	go func() {
		a.provider.Closer().CloseAll()
	}()

	a.provider.Closer().Wait()

	a.provider.Logger().Info("graceful shutdown completed")

	return nil
}

func (a *App) initDeps(ctx context.Context) error {
	a.initProvider(ctx)
	a.initDB(ctx)
	a.initGrpcServer(ctx)
	a.initHttpServer(ctx)

	return nil
}

func (a *App) initProvider(_ context.Context) {
	a.provider = NewProvider()
}

func (a *App) initDB(_ context.Context) {
	migrations.Run(a.provider.Cfg().Postgres)
}

func (a *App) initGrpcServer(ctx context.Context) {
	authInterceptor := interceptors.NewAuthInterceptor(
		a.provider.Parser(),
		routes.PublicRoutes(routes.GRPC),
	)

	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			interceptors.LoggerToCtxInterceptor(a.provider.Logger()),
			interceptors.RequestIDInterceptor,
			interceptors.LoggerInterceptor,
			authInterceptor.JWT,
		),
		grpc.Creds(insecure.NewCredentials()),
	}

	a.grpcServer = grpc.NewServer(opts...)

	reflection.Register(a.grpcServer)

	protogen.RegisterUsersServiceServer(a.grpcServer, a.provider.UsersHandler(ctx))

	a.provider.Closer().Add(func() error {
		a.grpcServer.GracefulStop()

		a.provider.Logger().Info("gRPC server stopped")
		return nil
	})
}

func (a *App) initHttpServer(ctx context.Context) {
	auth := middlewares.NewAuthMiddleware(a.provider.Parser(), routes.PublicRoutes(routes.HTTP))

	restServer := runtime.NewServeMux(
		runtime.WithMiddlewares(
			middlewares.LoggerToCtxInterceptor(a.provider.Logger()),
			middlewares.RequestIDInterceptor,
			middlewares.LoggerInterceptor,
			auth.JWT(),
		),
	)

	if err := protogen.RegisterUsersServiceHandlerServer(context.Background(), restServer, a.provider.UsersHandler(ctx)); err != nil {
		panic(err)
	}

	httpServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", a.provider.Cfg().HTTP.Host, a.provider.Cfg().HTTP.Port),
		Handler: restServer,
	}

	a.httpServer = httpServer

	a.provider.Closer().Add(func() error {
		if err := a.httpServer.Shutdown(ctx); err != nil {
			return err
		}

		a.provider.Logger().Info("http server stopped")

		return nil
	})
}

func (a *App) runGRPCServer(_ context.Context) error {
	a.provider.Logger().
		Info("starting gRPC server", slog.Int("port", a.provider.Cfg().GRPC.Port), slog.String("host", a.provider.Cfg().GRPC.Host))

	list, err := net.Listen(
		"tcp",
		fmt.Sprintf("%s:%d", a.provider.Cfg().GRPC.Host, a.provider.Cfg().GRPC.Port),
	)
	if err != nil {
		return fmt.Errorf("app.runGRPCServer: %w", err)
	}

	return a.grpcServer.Serve(list)
}

func (a *App) runHTTPServer(_ context.Context) error {
	a.provider.Logger().
		Info("starting HTTP server", slog.Int("port", a.provider.Cfg().HTTP.Port), slog.String("host", a.provider.Cfg().HTTP.Host))

	if err := a.httpServer.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}

		return fmt.Errorf("app.runHTTPServer: %w", err)
	}

	return nil
}
