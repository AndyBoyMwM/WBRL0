package nats

import (
	"log"
	"os"
	"prodWBRL0/internal/db"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
)

type ConnectionNatsHandler struct {
	conn  *stan.Conn
	subs  *Subscriber
	prod  *Producer
	name  string
	isErr bool
}

func NewNatsHandler(db *db.DB) *ConnectionNatsHandler {
	nhdlr := ConnectionNatsHandler{}
	nhdlr.initNatsHandlers(db)
	return &nhdlr
}

func (nhdlr *ConnectionNatsHandler) initNatsHandlers(db *db.DB) {
	nhdlr.name = "Nats"
	err := nhdlr.connectNats()

	if err != nil {
		nhdlr.isErr = true
		log.Printf("%s: connect nats error: %s", nhdlr.name, err)
	} else {
		nhdlr.subs = newSubscriber(db, nhdlr.conn)
		nhdlr.subs.Subscribe()

		nhdlr.prod = NewProducer(nhdlr.conn)
		nhdlr.prod.Produce()
	}
}

func (nhdlr *ConnectionNatsHandler) connectNats() error {
	conn, err := stan.Connect(
		os.Getenv("NATS_CLUSTER_ID"),
		os.Getenv("NATS_CLIENT_ID"),
		stan.NatsURL(os.Getenv("NATS_URL")),
		stan.NatsOptions(
			nats.ReconnectWait(time.Second*10),
			nats.Timeout(time.Second*10),
		),
		stan.Pings(5, 5),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			log.Printf("%s: connection lost, reason: %v", nhdlr.name, reason)
		}),
	)
	if err != nil {
		log.Printf("%s: connect nats error %v.\n", nhdlr.name, err)
		return err
	}
	nhdlr.conn = &conn

	log.Printf("%s: connect nats - OK", nhdlr.name)
	return nil
}

func (nhdlr *ConnectionNatsHandler) NatsUnsubscribeAndClose() {
	if !nhdlr.isErr {
		log.Printf("%s: unsubscribe & close...", nhdlr.name)
		nhdlr.subs.Unsubscribe()
		err := (*nhdlr.conn).Close()
		if err != nil {
			log.Printf("%s: unsubscribe & close error", nhdlr.name)
		}
		log.Printf("%s: unsubscribe & close OK", nhdlr.name)
	}
}
