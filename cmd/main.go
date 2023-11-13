package main

import (
	"fmt"
	"os"
	"os/signal"
	"prodWBRL0/cmd/config"
	"prodWBRL0/internal/db"
	"prodWBRL0/internal/db/cache"
	"prodWBRL0/internal/nats-streaming"
)

func main() {

	config.SetenvConfig()
	dbObject := db.NewDB()
	cache := cache.NewCache(dbObject)
	nhand := nats.NewNatsHandler(dbObject)
	chanelSignal := make(chan os.Signal, 1)
	chanelExit := make(chan bool)
	signal.Notify(chanelSignal, os.Interrupt)
	go func() {
		for range chanelSignal {
			cache.DeleteCache()
			nhand.NatsUnsubscribeAndClose()
			chanelExit <- true
			fmt.Printf("\nExit from main Goroutine\n\n\n")
		}
	}()
	<-chanelExit
}
