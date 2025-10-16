package clients

import (
	"context"
	"fmt"
	protoacc "gateway-service/api/protogen/accountservice/proto"
	"gateway-service/internal/config"
	"gateway-service/internal/logging"
	"gateway-service/internal/ports"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"sync"
	"time"
)

type GRPCAccountClient struct {
	conn    *grpc.ClientConn
	client  protoacc.AccountServiceClient
	mutex   sync.RWMutex
	timeout time.Duration
}

// NewAccountClient creating new grpc client
func NewAccountClient(timeout time.Duration) ports.AccountClient {
	return &GRPCAccountClient{
		timeout: timeout,
	}
}

func (c *GRPCAccountClient) Connect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.conn != nil {
		c.conn.Close()
	}

	_, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	conn, err := grpc.NewClient(
		config.Current().GRPC.AccountServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{
			MinConnectTimeout: c.timeout,
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to account service: %w", err)
	}

	c.conn = conn
	c.client = protoacc.NewAccountServiceClient(conn)
	logging.Logger.Info().Str("service", config.Current().GRPC.AccountServiceAddr).Msg("connected to account service")
	return nil
}

func (c *GRPCAccountClient) EnsureConnection() error {
	c.mutex.RLock()
	if c.conn != nil && c.conn.GetState() == connectivity.Ready {
		c.mutex.RUnlock()
		return nil
	}
	c.mutex.RUnlock()

	return c.Connect()
}

func (c *GRPCAccountClient) Close() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
		c.client = nil
	}
}

func (c *GRPCAccountClient) StartConnectionMonitor(ctx context.Context) {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := c.EnsureConnection(); err != nil {
				logging.Logger.Error().Err(err).Msg("Failed to maintain connection to account service")
			}
		}
	}
}

func (c *GRPCAccountClient) CreateCustomer(ctx context.Context, req *protoacc.CreateCustomerRequest) (*protoacc.CreateCustomerResponse, error) {
	if err := c.EnsureConnection(); err != nil {
		return nil, err
	}

	c.mutex.RLock()
	client := c.client
	c.mutex.RUnlock()

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return client.CreateCustomer(ctx, req)
}

func (c *GRPCAccountClient) UpdateCustomer(ctx context.Context, req *protoacc.UpdateCustomerRequest) (*protoacc.UpdateCustomerResponse, error) {
	if err := c.EnsureConnection(); err != nil {
		return nil, err
	}

	c.mutex.RLock()
	client := c.client
	c.mutex.RUnlock()

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return client.UpdateCustomer(ctx, req)
}

func (c *GRPCAccountClient) DeleteCustomer(ctx context.Context, req *protoacc.DeleteCustomerRequest) (*protoacc.DeleteCustomerResponse, error) {
	if err := c.EnsureConnection(); err != nil {
		return nil, err
	}

	c.mutex.RLock()
	client := c.client
	c.mutex.RUnlock()

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return client.DeleteCustomer(ctx, req)
}

func (c *GRPCAccountClient) GetCustomer(ctx context.Context, req *protoacc.GetCustomerRequest) (*protoacc.GetCustomerResponse, error) {
	if err := c.EnsureConnection(); err != nil {
		return nil, err
	}

	c.mutex.RLock()
	client := c.client
	c.mutex.RUnlock()

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return client.GetCustomer(ctx, req)
}

func (c *GRPCAccountClient) ListCustomer(ctx context.Context, req *protoacc.ListCustomersRequest) (*protoacc.ListCustomersResponse, error) {
	if err := c.EnsureConnection(); err != nil {
		return nil, err
	}

	c.mutex.RLock()
	client := c.client
	c.mutex.RUnlock()

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return client.ListCustomers(ctx, req)
}

func (c *GRPCAccountClient) CreateAccount(ctx context.Context, req *protoacc.CreateAccountRequest) (*protoacc.CreateAccountResponse, error) {
	if err := c.EnsureConnection(); err != nil {
		return nil, err
	}

	c.mutex.RLock()
	client := c.client
	c.mutex.RUnlock()

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return client.CreateAccount(ctx, req)
}

func (c *GRPCAccountClient) DeleteAccount(ctx context.Context, req *protoacc.DeleteAccountRequest) (*protoacc.DeleteAccountResponse, error) {
	if err := c.EnsureConnection(); err != nil {
		return nil, err
	}

	c.mutex.RLock()
	client := c.client
	c.mutex.RUnlock()

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return client.DeleteAccount(ctx, req)
}

func (c *GRPCAccountClient) ListAccount(ctx context.Context, req *protoacc.ListAccountsRequest) (*protoacc.ListAccountsResponse, error) {
	if err := c.EnsureConnection(); err != nil {
		return nil, err
	}

	c.mutex.RLock()
	client := c.client
	c.mutex.RUnlock()

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return client.ListAccount(ctx, req)
}

func (c *GRPCAccountClient) GetBalance(ctx context.Context, req *protoacc.GetBalanceRequest) (*protoacc.GetBalanceResponse, error) {
	if err := c.EnsureConnection(); err != nil {
		return nil, err
	}

	c.mutex.RLock()
	client := c.client
	c.mutex.RUnlock()

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return client.GetBalance(ctx, req)
}

func (c *GRPCAccountClient) InitTransaction(ctx context.Context, req *protoacc.InitTransactionRequest) (*protoacc.InitTransactionResponse, error) {
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

func (c *GRPCAccountClient) GetTransactionHistory(ctx context.Context, req *protoacc.GetTransactionHistoryRequest) (*protoacc.GetTransactionHistoryResponse, error) {
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
