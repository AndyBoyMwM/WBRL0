package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"os"
)

func (db *DB) Init() {
	db.name = "Postgres"
	var err error
	dsn := fmt.Sprintf("postgres://%v:%v@%v/%v", os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_HOST"), os.Getenv("DB_NAME"))
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatalf("%v: config error: %s\n", db.name, err)
	}
	db.pool, err = pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		log.Fatalf("%v: connect  error: %v\n", db.name, err)
	}
	log.Printf("%v: connect - OK\n", db.name)
}
