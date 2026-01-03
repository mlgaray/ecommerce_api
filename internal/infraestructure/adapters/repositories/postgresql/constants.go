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
	ScanField              = "scan"
	UnmarshallField        = "unmarshall"
	NextField              = "next"
)

// Sort order constants
const (
	SortOrderDesc = "desc"
)

// PostgreSQL error codes
const (
	PqErrCodeUniqueViolation     = "23505"
	PqErrCodeForeignKeyViolation = "23503"
	PqErrCodeRaiseException      = "P0003"
	PqErrCodeNoDataFound         = "P0002"
)

// Database constraint names
const (
	ConstraintCategoryNameShopIDUnique = "categories_name_shop_id_unique"
)

// JSON values
const (
	JSONNull = "null"
)
