package grpc

import (
	"net"

	"github.com/marcelbeumer/go-playground/gochat/internal/chat"
	"github.com/marcelbeumer/go-playground/gochat/internal/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type HubService struct {
	logger log.Logger
	hub    chat.Hub
}

func (h *HubService) Chat(s Hub_ChatServer) error {
	md, ok := metadata.FromIncomingContext(s.Context())
	if !ok {
		return status.Error(codes.FailedPrecondition, "no metadata found")
	}
	usernames := md.Get("username")
	if len(usernames) == 0 || usernames[0] == "" {
		return status.Error(codes.FailedPrecondition, "no username in metadata")
	}
	username := usernames[0]
	conn := NewConnection(s, h.logger)
	_, err := h.hub.Connect(username, conn)
	if err != nil {
		return status.Error(codes.Unknown, err.Error())
	}
	err = conn.Wait()
	if err != nil {
		return status.Error(codes.Unknown, err.Error())
	}
	return nil
}

func (h *HubService) mustEmbedUnimplementedHubServer() {}

type Server struct {
	logger     log.Logger
	grpcServer *grpc.Server
}

func (s *Server) Start(addr string) error {
	logger := s.logger
	logger.Infow("starting grpc server", "addr", addr)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	var opts []grpc.ServerOption
	s.grpcServer = grpc.NewServer(opts...)

	RegisterHubServer(s.grpcServer, &HubService{
		logger: s.logger,
		hub:    *chat.NewHub(logger),
	})

	_ = s.grpcServer.Serve(lis)
	return err
}

func (s *Server) Stop() error {
	if s.grpcServer != nil {
		s.grpcServer.Stop()
	}
	return nil
}

func NewServer(logger log.Logger) *Server {
	return &Server{
		logger: logger,
	}
}
