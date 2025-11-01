package postgresql

// Infrastructure-level log message constants for PostgreSQL operations
const (
	// Transaction operations
	FailedBeginTransactionLog  = "Failed to begin transaction"
	FailedCommitTransactionLog = "Failed to commit transaction"

	// General database operations
	DatabaseQueryFailedLog = "Database query failed"
	DatabaseScanFailedLog  = "Database scan failed"

	BeginTransactionField  = "begin_transaction"
	CommitTransactionField = "commit_transaction"
	ScanField       ="scan"
	UnmarshallField ="unmarshall"
	NextField="next"
)
