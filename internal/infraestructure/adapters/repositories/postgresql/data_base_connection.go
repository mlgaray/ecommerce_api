package postgresql

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type DataBaseConnection interface {
	Connect() *sql.DB
}

type dataBaseConnection struct{}

func (c *dataBaseConnection) Connect() *sql.DB {
	// c.envService.LoadEnv()

	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	dataSourceName := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable", dbUser, dbPassword, dbHost, dbPort, dbName)

	// Fixed: first parameter should be "postgres", not dbName
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		log.Fatal(err)
	}
	// defer db.Close()

	// CONNECTION POOL OPTIMIZATION
	// MaxOpenConns: Hard ceiling on total connections (in-use + idle)
	// Recommended: Based on DB capacity and number of app replicas
	db.SetMaxOpenConns(25)

	// MaxIdleConns: Should equal MaxOpenConns to avoid frequent reconnections
	// When smaller than MaxOpenConns, connections open/close more frequently
	db.SetMaxIdleConns(25)

	// ConnMaxLifetime: Recycle connections periodically for reliability
	// Helps with load balancers and prevents stale connections
	db.SetConnMaxLifetime(5 * time.Minute)

	// ConnMaxIdleTime: Close idle connections after period of inactivity
	// Allows pool to scale down after high load periods
	db.SetConnMaxIdleTime(5 * time.Minute)

	// Verifica la conexión
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Conexión exitosa a la base de datos!")
	fmt.Printf("Connection pool configured: MaxOpen=%d, MaxIdle=%d\n", 25, 25)

	return db
}

func NewDataBaseConnection() *dataBaseConnection {
	return &dataBaseConnection{}
}
