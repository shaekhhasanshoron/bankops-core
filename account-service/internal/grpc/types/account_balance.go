package types

type AccountBalance struct {
	AccountID string
	Balance   float64
	Version   int
}

type AccountBalanceResponse struct {
	AccountID string
	Version   int
}
