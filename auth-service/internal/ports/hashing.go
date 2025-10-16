package ports

// Hashing is responsible for signing hashing Data tokens.
type Hashing interface {
	HashData(data string) (string, error)
}
