package clients

import (
	"context"
	"fmt"
	prototx "gateway-service/api/protogen/txservice/proto"
	"gateway-service/internal/logging"
	"gateway-service/internal/ports"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"sync"
	"time"
)

type GRPCTransactionClient struct {
	conn     *grpc.ClientConn
	client   prototx.TransactionServiceClient
	mutex    sync.RWMutex
	timeout  time.Duration
	grpcAddr string
}

// NewTransactionClient creating new grpc client
func NewTransactionClient(timeout time.Duration, grpcAddr string) ports.TransactionClient {
	return &GRPCTransactionClient{
		timeout:  timeout,
		grpcAddr: grpcAddr,
	}
}

func (c *GRPCTransactionClient) Connect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.conn != nil {
		c.conn.Close()
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	conn, err := grpc.NewClient(
		c.grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{
			MinConnectTimeout: c.timeout,
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to transaction service: %w", err)
	}

	c.conn = conn
	c.client = prototx.NewTransactionServiceClient(conn)
	resp, err := c.client.HealthCheck(ctx, &prototx.HealthCheckRequest{
		Message: "ping",
	})
	if err != nil || resp.Message != "pong" {
		_ = c.conn.Close()
		logging.Logger.Warn().Err(err).Str("service", c.grpcAddr).
			Msg("failed to connect to transaction service. Retrying to connect...")
	} else {
		logging.Logger.Info().Str("service", c.grpcAddr).Msg("connected to transaction service")
	}
	return nil
}

func (c *GRPCTransactionClient) EnsureConnection() error {
	c.mutex.RLock()
	if c.conn != nil && c.conn.GetState() == connectivity.Ready {
		c.mutex.RUnlock()
		return nil
	}
	c.mutex.RUnlock()

	return c.Connect()
}

func (c *GRPCTransactionClient) IsHealthy() bool {
	c.mutex.RLock()
	if c.conn != nil && c.conn.GetState() == connectivity.Ready {
		c.mutex.RUnlock()
		return true
	}
	return false
}

func (c *GRPCTransactionClient) Close() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
		c.client = nil
	}
}

func (c *GRPCTransactionClient) StartConnectionMonitor(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := c.EnsureConnection(); err != nil {
				logging.Logger.Error().Err(err).Msg("Failed to maintain connection to transaction service")
			}
		}
	}
}

func (c *GRPCTransactionClient) InitTransaction(ctx context.Context, req *prototx.InitTransactionRequest) (*prototx.InitTransactionResponse, error) {
	if err := c.EnsureConnection(); err != nil {
		return nil, err
	}

	c.mutex.RLock()
	client := c.client
	c.mutex.RUnlock()

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return client.InitTransaction(ctx, req)
}

func (c *GRPCTransactionClient) GetTransactionHistory(ctx context.Context, req *prototx.GetTransactionHistoryRequest) (*prototx.GetTransactionHistoryResponse, error) {
	if err := c.EnsureConnection(); err != nil {
		return nil, err
	}

	c.mutex.RLock()
	client := c.client
	c.mutex.RUnlock()

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return client.GetTransactionHistory(ctx, req)
}
