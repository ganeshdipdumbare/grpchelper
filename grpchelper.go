package grpchelper

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Server holds grpc server and tcp listner
type Server struct {
	GrpcServer *grpc.Server
	Listner    *net.Listener
}

// NewServer creates grpc server with given configuration and tcp listner for given address
func NewServer(addr string, ui []grpc.UnaryServerInterceptor, si []grpc.StreamServerInterceptor, enableReflection bool) (*Server, error) {
	grpcServer := setupServer(ui, si, enableReflection)

	listner, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &Server{
		GrpcServer: grpcServer,
		Listner:    &listner,
	}, nil
}

func setupServer(ui []grpc.UnaryServerInterceptor, si []grpc.StreamServerInterceptor, enableReflection bool) *grpc.Server {
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			ui...,
		)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			si...,
		)),
	)

	if enableReflection {
		reflection.Register(grpcServer)
	}

	return grpcServer
}

// Serve starts TCP server and will keep running
func (s *Server) Serve() error {
	err := s.GrpcServer.Serve(*s.Listner)
	if err != nil {
		return err
	}
	return nil
}

// AwaitTermination waits for termination signal and when it is received
// gracefully stops the grpc server and close the tcp listner
func (s *Server) AwaitTermination() {

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	defer signal.Stop(stop)
	<-stop

	s.GrpcServer.GracefulStop()
	(*s.Listner).Close()
}
