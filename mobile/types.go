package mobile

// APIKeyResult holds the result of API key generation
type APIKeyResult struct {
	PrivateKey string
	PublicKey  string
	Error      string
}

// TxResult holds the result of a transaction signing operation
type TxResult struct {
	JSON  string
	Error string
}

