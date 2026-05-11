package sender

import (
	"context"
	"fmt"
	"time"

	bangmodv1 "github.com/bangmodmonitor/gen/bangmod/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type GRPCSender struct {
	token  string
	client bangmodv1.MetricsServiceClient
	conn   *grpc.ClientConn
}

func NewGRPC(target, token string) (*GRPCSender, error) {
	conn, err := grpc.NewClient(target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                30 * time.Second,
			Timeout:             10 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(4*1024*1024)),
	)
	if err != nil {
		return nil, fmt.Errorf("grpc dial %s: %w", target, err)
	}
	return &GRPCSender{
		token:  token,
		client: bangmodv1.NewMetricsServiceClient(conn),
		conn:   conn,
	}, nil
}

func (s *GRPCSender) Send(metrics *bangmodv1.AgentMetrics) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := s.client.IngestMetrics(ctx, &bangmodv1.IngestRequest{
		Token:   s.token,
		Metrics: metrics,
	})
	if err != nil {
		return err
	}
	if !resp.Ok {
		return fmt.Errorf("server rejected: %s", resp.Message)
	}
	return nil
}

func (s *GRPCSender) Close() error {
	return s.conn.Close()
}
