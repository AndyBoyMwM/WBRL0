package nats

import (
	"encoding/json"
	"github.com/nats-io/stan.go"
	"log"
	"math/rand"
	"os"
	"prodWBRL0/internal/db/models"
	"strconv"
	"time"
)

type Producer struct {
	conn *stan.Conn
	name string
}

func NewProducer(conn *stan.Conn) *Producer {
	return &Producer{
		name: "Producer",
		conn: conn,
	}
}

func (p *Producer) Produce() {

	rand.Seed(time.Now().UnixNano())
	trackNumber := strconv.Itoa(rand.Intn(9)) + "WBILM" + strconv.Itoa(rand.Intn(10000))
	orderUID := strconv.Itoa(rand.Intn(100)) + "order" + strconv.Itoa(rand.Intn(10000))
	sA, sB := rand.Intn(5)+1, rand.Intn(3)+1
	pA, pB := rand.Intn(5000), rand.Intn(10000)

	itemA := models.Items{ChrtID: 100, TrackNumber: trackNumber, Price: pA, Rid: "ab4219087a764ae0w5dcz1",
		Name: "Mascaras", Sale: sA, Size: "XS", TotalPrice: sA * pA, NmID: rand.Intn(10000000) + 1, Brand: "Vivienne_Sabo", Status: 200}

	itemB := models.Items{ChrtID: 33, TrackNumber: trackNumber, Price: pB, Rid: "123efwfwef2332342efsdv",
		Name: "Taifun", Sale: sB, Size: "XX", TotalPrice: sB * pB, NmID: 4234412, Brand: "Gerry_Weber", Status: 230}

	payment := models.Payment{Transaction: orderUID, RequestID: "", Currency: "RUB", Provider: "wbpay", Amount: sA*pA + sB*pB,
		PaymentDt: 1637907727, Bank: "alpha", DeliveryCost: 5000, GoodsTotal: 7}

	delivery := models.Delivery{Name: "Best delivery", Phone: "+9720000000", Zip: "2639809", City: "Best city",
		Address: "Ploshad Mira 15", Region: "Kraiot", Email: "test@gmail.com"}

	order := models.Order{OrderUID: orderUID, TrackNumber: trackNumber, Entry: "WBIL",
		Delivery: delivery, Payment: payment, Items: []models.Items{itemA, itemB},
		Locale: "Ru", InternalSignature: "", CustomerID: "BestCustomer", DeliveryService: "fura",
		ShardKey: "SHK1", SmID: 99, OofShard: "1"}
	orderData, err := json.Marshal(order)
	if err != nil {
		log.Printf("%s: json.Marshal(order) err: %v\n", p.name, err)
	}

	// Asynchronous Publishing
	ackHandler := func(ackedNuid string, err error) {
		if err != nil {
			log.Printf("%s: error publishing msg id %s: %v\n", p.name, ackedNuid, err.Error())
		} else {
			log.Printf("%s: received ack for msg id: %s\n", p.name, ackedNuid)
		}
	}

	log.Printf("%s: pblish async NATS_SUBJECT...\n", p.name)
	nuid, err := (*p.conn).PublishAsync(os.Getenv("NATS_SUBJECT"), orderData, ackHandler) // returns immediately
	if err != nil {
		log.Printf("%s: error pulish async msg %s: %v\n", p.name, nuid, err.Error())
	}
}
