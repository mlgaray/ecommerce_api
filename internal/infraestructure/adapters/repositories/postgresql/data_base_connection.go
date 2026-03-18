package postgresql

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/metrics"
)

type DataBaseConnection interface {
	Connect() *sql.DB
}

type dataBaseConnection struct {
	once sync.Once
	db   *sql.DB
}

func (c *dataBaseConnection) Connect() *sql.DB {
	c.once.Do(func() {
		dbUser := os.Getenv("DB_USER")
		dbPassword := os.Getenv("DB_PASSWORD")
		dbHost := os.Getenv("DB_HOST")
		dbPort := os.Getenv("DB_PORT")
		dbName := os.Getenv("DB_NAME")
		sslMode := os.Getenv("DB_SSLMODE")
		if sslMode == "" {
			sslMode = "require"
		}
		// Use URL format — required for Supabase pooler where username contains a dot
		dataSourceName := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=%s", dbUser, dbPassword, dbHost, dbPort, dbName, sslMode)

		// Register the instrumented driver wrapping lib/pq for query-level metrics
		driverName := metrics.RegisterInstrumentedDriver()

		db, err := sql.Open(driverName, dataSourceName)
		if err != nil {
			log.Fatal(err)
		}

		// CONNECTION POOL OPTIMIZATION
		// MaxOpenConns: Hard ceiling on total connections (in-use + idle)
		db.SetMaxOpenConns(25)

		// MaxIdleConns: Should equal MaxOpenConns to avoid frequent reconnections
		db.SetMaxIdleConns(25)

		// ConnMaxLifetime: Recycle connections periodically for reliability
		db.SetConnMaxLifetime(5 * time.Minute)

		// ConnMaxIdleTime: Close idle connections after period of inactivity
		db.SetConnMaxIdleTime(5 * time.Minute)

		err = db.Ping()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Conexión exitosa a la base de datos!")
		fmt.Printf("Connection pool configured: MaxOpen=%d, MaxIdle=%d\n", 25, 25)

		c.db = db
	})

	return c.db
}

func NewDataBaseConnection() *dataBaseConnection {
	return &dataBaseConnection{}
}
