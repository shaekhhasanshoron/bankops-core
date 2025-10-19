package clients

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"sync"
	"time"
	protoacc "transaction-service/api/protogen/accountservice/proto"
	"transaction-service/internal/config"
	"transaction-service/internal/logging"
	"transaction-service/internal/ports"
)

type GRPCAccountClient struct {
	conn     *grpc.ClientConn
	client   protoacc.AccountServiceClient
	mutex    sync.RWMutex
	timeout  time.Duration
	grpcAddr string
}

func NewAccountClient(timeout time.Duration, grpcAddr string) ports.AccountClient {
	return &GRPCAccountClient{
		timeout:  timeout,
		grpcAddr: grpcAddr,
	}
}

func (c *GRPCAccountClient) Connect() error {
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
		return fmt.Errorf("failed to connect to account service: %w", err)
	}

	c.conn = conn
	c.client = protoacc.NewAccountServiceClient(conn)
	resp, err := c.client.HealthCheck(ctx, &protoacc.HealthCheckRequest{
		Message: "ping",
	})
	if err != nil || resp.Message != "pong" {
		_ = c.conn.Close()
		logging.Logger.Warn().Err(err).Str("service", config.Current().GRPC.AccountServiceAddr).
			Msg("failed to connect to account service. Retrying to connect...")
	} else {
		logging.Logger.Info().Str("service", config.Current().GRPC.AccountServiceAddr).Msg("connected to account service")
	}
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

func (c *GRPCAccountClient) ValidateAndGetAccounts(ctx context.Context, accountIDs []string, requester, requestId string) ([]ports.AccountInfo, string, error) {
	if c.IsHealthy() == false {
		return nil, "connection failed", fmt.Errorf("connection failed")
	}

	c.mutex.RLock()
	client := c.client
	c.mutex.RUnlock()

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	req := &protoacc.ValidateAccountsRequest{
		AccountIds: accountIDs,
		Metadata: &protoacc.Metadata{
			RequestId: requester,
			Requester: requestId,
		},
	}

	resp, err := client.ValidateAccounts(ctx, req)
	if err != nil {
		return nil, "validate accounts RPC failed", err
	}

	if !resp.Valid {
		return nil, resp.Message, fmt.Errorf("account validation failed")
	}

	var accounts []ports.AccountInfo
	for _, account := range resp.Accounts {
		accounts = append(accounts, ports.AccountInfo{
			AccountID:  account.Id,
			CustomerID: account.CustomerId,
			Balance:    account.Balance,
			Version:    int(account.Version),
		})
	}

	return accounts, "", nil
}

func (c *GRPCAccountClient) LockAccounts(ctx context.Context, accountIDs []string, transactionID string, requester, requestId string) (string, error) {
	if c.IsHealthy() == false {
		return "connection failed", fmt.Errorf("connection failed")
	}

	c.mutex.RLock()
	client := c.client
	c.mutex.RUnlock()

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	req := &protoacc.LockAccountsRequest{
		AccountIds:    accountIDs,
		TransactionId: transactionID,
		Metadata: &protoacc.Metadata{
			RequestId: requester,
			Requester: requestId,
		},
	}

	resp, err := client.LockAccounts(ctx, req)
	if err != nil {
		return "lock accounts RPC failed", err
	}

	if !resp.Locked {
		return resp.Message, fmt.Errorf("account locking failed")
	}

	return "", nil
}

func (c *GRPCAccountClient) UnlockAccounts(ctx context.Context, transactionID string, requester, requestId string) (string, error) {
	if c.IsHealthy() == false {
		return "connection failed", fmt.Errorf("connection failed")
	}

	c.mutex.RLock()
	client := c.client
	c.mutex.RUnlock()

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	req := &protoacc.UnlockAccountsRequest{
		TransactionId: transactionID,
		Metadata: &protoacc.Metadata{
			RequestId: requester,
			Requester: requestId,
		},
	}

	resp, err := client.UnlockAccounts(ctx, req)
	if err != nil {
		return "unlock accounts RPC failed", err
	}

	if !resp.Unlocked {
		return resp.Message, fmt.Errorf("account unlocking failed")
	}

	return "", nil
}

func (c *GRPCAccountClient) UpdateAccountsBalance(ctx context.Context, updates []ports.AccountBalanceUpdate, requester, requestId string) ([]ports.AccountBalanceUpdateResponse, string, error) {
	if c.IsHealthy() == false {
		return nil, "connection failed", fmt.Errorf("connection failed")
	}

	c.mutex.RLock()
	client := c.client
	c.mutex.RUnlock()

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	protoUpdates := make([]*protoacc.AccountBalanceUpdate, len(updates))
	for i, update := range updates {
		protoUpdates[i] = &protoacc.AccountBalanceUpdate{
			AccountId:  update.AccountID,
			NewBalance: update.NewBalance,
			Version:    int32(update.Version),
		}
	}

	req := &protoacc.UpdateAccountsBalanceRequest{
		Updates: protoUpdates,
		Metadata: &protoacc.Metadata{
			RequestId: requester,
			Requester: requestId,
		},
	}

	resp, err := client.UpdateAccountsBalance(ctx, req)
	if err != nil {
		return nil, "update accounts balance RPC failed", err
	}

	if !resp.Success || resp.NewVersions == nil || len(resp.NewVersions) == 0 {
		return nil, resp.Message, fmt.Errorf("balance update failed")
	}

	var accountVersionResponse []ports.AccountBalanceUpdateResponse
	for _, rsp := range resp.NewVersions {
		accountVersionResponse = append(accountVersionResponse, ports.AccountBalanceUpdateResponse{
			AccountID: rsp.AccountId,
			Version:   int(rsp.Version),
		})
	}
	return accountVersionResponse, "", nil
}

func (c *GRPCAccountClient) GetBalance(ctx context.Context, accountID string) (float64, int, error) {
	if c.IsHealthy() == false {
		return 0, 0, fmt.Errorf("connection failed")
	}

	c.mutex.RLock()
	client := c.client
	c.mutex.RUnlock()

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	req := &protoacc.GetBalanceRequest{
		AccountId: accountID,
	}

	resp, err := client.GetBalance(ctx, req)
	if err != nil {
		return 0, 0, fmt.Errorf("get balance RPC failed: %w", err)
	}

	return resp.Balance, int(resp.Version), nil
}

func (c *GRPCAccountClient) IsHealthy() bool {
	c.mutex.RLock()
	if c.conn != nil && c.conn.GetState() == connectivity.Ready {
		c.mutex.RUnlock()
		return true
	}
	return false
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
	ticker := time.NewTicker(5 * time.Second)
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
