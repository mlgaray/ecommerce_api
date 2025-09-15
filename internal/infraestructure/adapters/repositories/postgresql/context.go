package postgresql

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const TxContextKey contextKey = "tx"
