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
	// MaxOpenConns: Maximum number of open connections to the database
	// Recommended: 2-4x number of CPU cores for I/O bound operations
	db.SetMaxOpenConns(25)

	// MaxIdleConns: Maximum number of connections in the idle connection pool
	// Should be less than or equal to MaxOpenConns
	// Higher value = faster query execution (no need to create new connections)
	db.SetMaxIdleConns(10)

	// ConnMaxLifetime: Maximum amount of time a connection may be reused
	// Prevents issues with stale connections and database connection limits
	db.SetConnMaxLifetime(5 * time.Minute)

	// ConnMaxIdleTime: Maximum amount of time a connection may be idle before being closed
	// Helps free up resources when load is low
	db.SetConnMaxIdleTime(1 * time.Minute)

	// Verifica la conexión
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Conexión exitosa a la base de datos!")
	fmt.Printf("Connection pool configured: MaxOpen=%d, MaxIdle=%d\n", 25, 10)

	return db
}

func NewDataBaseConnection() *dataBaseConnection {
	return &dataBaseConnection{}
}
