package grpc

import (
	"context"

	"github.com/marcelbeumer/crispy-octo-goggles/gochat/internal/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func NewClientConnection(
	serverAddr string,
	username string,
	logger log.Logger,
) (*Connection, error) {
	logger.Infow("connecting to server", "serverUrl", serverAddr)
	conn, err := grpc.Dial(
		serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	client := NewHubClient(conn)
	header := metadata.New(map[string]string{"username": username})
	ctx := metadata.NewOutgoingContext(context.Background(), header)
	cc, err := client.Chat(ctx)
	if err != nil {
		return nil, err
	}
	return NewConnection(cc, logger), nil
}
