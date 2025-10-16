package clients

import (
	"context"
	"fmt"
	protoauth "gateway-service/api/protogen/authservice/proto"
	"gateway-service/internal/config"
	"gateway-service/internal/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"sync"
	"time"
)

type GRPCAuthClient struct {
	conn    *grpc.ClientConn
	client  protoauth.AuthServiceClient
	mutex   sync.RWMutex
	timeout time.Duration
}

// NewAuthClient creating new grpc client
func NewAuthClient(timeout time.Duration) AuthClient {
	return &GRPCAuthClient{
		timeout: timeout,
	}
}

func (c *GRPCAuthClient) Connect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.conn != nil {
		c.conn.Close()
	}

	_, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	conn, err := grpc.NewClient(
		config.Current().GRPC.AuthServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{
			MinConnectTimeout: c.timeout,
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to auth service: %w", err)
	}

	c.conn = conn
	c.client = protoauth.NewAuthServiceClient(conn)
	logging.Logger.Info().Str("service", config.Current().GRPC.AuthServiceAddr).Msg("connected to auth service")
	return nil
}

func (c *GRPCAuthClient) EnsureConnection() error {
	c.mutex.RLock()
	if c.conn != nil && c.conn.GetState() == connectivity.Ready {
		c.mutex.RUnlock()
		return nil
	}
	c.mutex.RUnlock()

	return c.Connect()
}

func (c *GRPCAuthClient) Authenticate(ctx context.Context, req *protoauth.AuthenticateRequest) (*protoauth.AuthenticateResponse, error) {
	if err := c.EnsureConnection(); err != nil {
		return nil, err
	}

	c.mutex.RLock()
	client := c.client
	c.mutex.RUnlock()

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return client.Authenticate(ctx, req)
}

func (c *GRPCAuthClient) CreateEmployee(ctx context.Context, req *protoauth.CreateEmployeeRequest) (*protoauth.CreateEmployeeResponse, error) {
	if err := c.EnsureConnection(); err != nil {
		return nil, err
	}

	c.mutex.RLock()
	client := c.client
	c.mutex.RUnlock()

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return client.CreateEmployee(ctx, req)
}

func (c *GRPCAuthClient) DeleteEmployee(ctx context.Context, req *protoauth.DeleteEmployeeRequest) (*protoauth.DeleteEmployeeResponse, error) {
	if err := c.EnsureConnection(); err != nil {
		return nil, err
	}

	c.mutex.RLock()
	client := c.client
	c.mutex.RUnlock()

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return client.DeleteEmployee(ctx, req)
}

func (c *GRPCAuthClient) UpdateEmployee(ctx context.Context, req *protoauth.UpdateRoleRequest) (*protoauth.UpdateRoleResponse, error) {
	if err := c.EnsureConnection(); err != nil {
		return nil, err
	}

	c.mutex.RLock()
	client := c.client
	c.mutex.RUnlock()

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return client.UpdateRole(ctx, req)
}

func (c *GRPCAuthClient) ListEmployee(ctx context.Context, req *protoauth.ListEmployeeRequest) (*protoauth.ListEmployeeResponse, error) {
	if err := c.EnsureConnection(); err != nil {
		return nil, err
	}

	c.mutex.RLock()
	client := c.client
	c.mutex.RUnlock()

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return client.ListEmployee(ctx, req)
}

func (c *GRPCAuthClient) Close() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
		c.client = nil
	}
}

func (c *GRPCAuthClient) StartConnectionMonitor(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := c.EnsureConnection(); err != nil {
				logging.Logger.Error().Err(err).Msg("Failed to maintain connection to auth service")
			}
		}
	}
}
