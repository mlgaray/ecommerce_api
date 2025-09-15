package postgresql

import (
	"database/sql"
	"fmt"
	"log"
	"os"

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
	dataSourceName := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s", dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err := sql.Open(dbName, dataSourceName)
	if err != nil {
		log.Fatal(err)
	}
	// defer db.Close()

	// Verifica la conexión
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Conexión exitosa a la base de datos!")

	return db
}

func NewDataBaseConnection() *dataBaseConnection {
	return &dataBaseConnection{}
}
