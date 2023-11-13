package nats

import (
	"encoding/json"
	"github.com/nats-io/stan.go"
	"log"
	"os"
	"prodWBRL0/internal/db"
	"prodWBRL0/internal/db/models"
	"strconv"
	"time"
)

type Subscriber struct {
	subs  stan.Subscription
	dbObj *db.DB
	stc   *stan.Conn
	name  string
}

func newSubscriber(db *db.DB, conn *stan.Conn) *Subscriber {
	return &Subscriber{
		name:  "Subscriber",
		dbObj: db,
		stc:   conn,
	}
}

func (s *Subscriber) Subscribe() {

	natsIdleTimout, err := strconv.Atoi(os.Getenv("NATS_IDLE_TIMEOUT"))
	if err != nil {
		log.Printf("%s: got the message!\n", s.name)
		return
	}

	s.subs, err = (*s.stc).Subscribe(
		os.Getenv("NATS_SUBJECT"),
		func(m *stan.Msg) {
			log.Printf("%s: got a new message!\n", s.name)
			if s.messageHandler(m.Data) {
				err := m.Ack()
				if err != nil {
					log.Printf("%s ack() err: %s", s.name, err)
				}
			}
		},
		stan.AckWait(time.Duration(natsIdleTimout)*time.Second),
		stan.DurableName(os.Getenv("NATS_DURABLE_NAME")),
		stan.SetManualAckMode(),
		stan.MaxInflight(5))
	if err != nil {
		log.Printf("%s: error: %v\n", s.name, err)
	}
	log.Printf("%s: subscribe to NATS_SUBJECT %s\n", s.name, os.Getenv("NATS_SUBJECT"))
}

func (s *Subscriber) messageHandler(data []byte) bool {
	order := models.Order{}
	err := json.Unmarshal(data, &order)
	if err != nil {
		log.Printf("%s: unmarshal order error, %v\n", s.name, err)
		return true
	}
	log.Printf("%s: unmarshal order - OK: %v\n", s.name, order)

	_, err = s.dbObj.AddOrder(order)
	if err != nil {
		log.Printf("%s:can`t insert order to DB: %v\n", s.name, err)
		return false
	}
	return true
}

func (s *Subscriber) Unsubscribe() {
	if s.subs != nil {
		err := s.subs.Unsubscribe()
		if err != nil {
			log.Printf("%s: can`t unsubscribe: %v\n", s.name, err)
		}
	}
}
