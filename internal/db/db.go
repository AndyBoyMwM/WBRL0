package db

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"prodWBRL0/internal/db/cache"
	"prodWBRL0/internal/db/models"

	"github.com/jackc/pgx/v4/pgxpool"
)

type DB struct {
	pool *pgxpool.Pool
	csh  *cache.Cache
	name string
}

func NewDB() *DB {
	db := DB{}
	db.Init()
	return &db
}

func (db *DB) SetCacheInstance(csh *cache.Cache) {
	db.csh = csh
}

func (db *DB) GetCacheState(bufSize int) (map[int64]models.Order, []int64, int, error) {
	buffer := make(map[int64]models.Order, bufSize)
	priority := make([]int64, bufSize)
	var ind int

	query := fmt.Sprintf("SELECT order_id FROM cache WHERE cache_stream = '%s' ORDER BY id DESC LIMIT %d",
		os.Getenv("DB_CACHE_STREAM"), bufSize)
	rows, err := db.pool.Query(context.Background(), query)
	if err != nil {
		log.Printf("%v: can`t select order_id from cache table: %v\n", db.name, err)
	}
	log.Printf("%v: rows`cache: %v\n", db.name, rows)
	defer rows.Close()

	var orderid int64
	for rows.Next() {
		if err := rows.Scan(&orderid); err != nil {
			log.Printf("%v: can`t select order_id from cache table: %v\n", db.name, err)
			return buffer, priority, ind, errors.New("can`t select order_id from cache table")
		}
		log.Printf("%v:`cache orderid: %v\n", db.name, orderid)
		priority[ind] = orderid
		ind++

		o, err := db.GetOrderByID(orderid)
		if err != nil {
			log.Printf("%v: can`t select order from DB %v\n", db.name, err)
			continue
		}
		buffer[orderid] = o
	}

	if ind == 0 {
		return buffer, priority, ind, errors.New("cache index = 0")
	}

	for i := 0; i < int(ind/2); i++ {
		priority[i], priority[ind-i-1] = priority[ind-i-1], priority[i]
	}

	return buffer, priority, ind, nil
}

func (db *DB) GetOrderByID(oid int64) (models.Order, error) {
	var o models.Order
	var paymentIdFk, deliveryIdFk int64

	err := db.pool.QueryRow(context.Background(), `SELECT order_uid, track_number, entry, delivery_id_fk, 
        payment_id_fk, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
		FROM orders WHERE id = $1`, oid).Scan(&o.OrderUID, &o.TrackNumber, &o.Entry, &deliveryIdFk, &paymentIdFk, &o.Locale,
		&o.InternalSignature, &o.CustomerID, &o.DeliveryService, &o.ShardKey, &o.SmID, &o.DateCreated, &o.OofShard)
	if err != nil {
		return o, errors.New("can`t select from orders table")
	}

	err = db.pool.QueryRow(context.Background(), `SELECT transaction, request_id, currency, provider, amount, 
    	payment_dt, bank, delivery_cost, goods_total, custom_fee FROM payments WHERE id = $1`, paymentIdFk).
		Scan(&o.Payment.Transaction, o.Payment.RequestID, &o.Payment.Currency, &o.Payment.Provider, &o.Payment.Amount,
			&o.Payment.PaymentDt, &o.Payment.Bank, &o.Payment.DeliveryCost, &o.Payment.GoodsTotal)
	if err != nil {
		log.Printf("%v: unable to get payment from database: %v\n", db.name, err)
		return o, errors.New("can`t select from payments table")
	}

	err = db.pool.QueryRow(context.Background(), `SELECT name, phone, zip, city, address, region, email 
		FROM deliverys WHERE id = $1`, deliveryIdFk).Scan(&o.Delivery.Name, o.Delivery.Phone, &o.Delivery.Zip,
		&o.Delivery.City, &o.Delivery.Address, &o.Delivery.Region, &o.Delivery.Email)
	if err != nil {
		log.Printf("%v:can`t select from deliverys table: %v\n", db.name, err)
		return o, errors.New("can`t select from deliverys table")
	}

	rowsItems, err := db.pool.Query(context.Background(), "SELECT item_id_fk FROM order_items WHERE order_id_fk = $1", oid)
	if err != nil {
		return o, errors.New("can`t select item_id_fk from order_items table")
	}
	defer rowsItems.Close()

	var itemID int64
	for rowsItems.Next() {
		var i models.Items
		if err := rowsItems.Scan(&itemID); err != nil {
			return o, errors.New("can`t get item_id_fk from rowsItems[]")
		}

		err = db.pool.QueryRow(context.Background(), `SELECT ChrtID, Price, Rid, Name, Sale, Size, TotalPrice, NmID, Brand 
		FROM items WHERE id = $1`, itemID).Scan(&i.ChrtID, &i.Price, &i.Rid, &i.Name, &i.Sale, &i.Size,
			&i.TotalPrice, &i.NmID, &i.Brand)
		if err != nil {
			return o, errors.New("can`t select item from items table")
		}
		o.Items = append(o.Items, i)
	}
	return o, nil
}

func (db *DB) AddOrder(o models.Order) (int64, error) {
	var lastInsertId int64
	var itemsIds = []int64{}

	tx, err := db.pool.Begin(context.Background())
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(context.Background())

	for _, item := range o.Items {
		err := tx.QueryRow(context.Background(), `INSERT INTO items (chrt_id, track_number, price, rid, name, sale,
		size, total_price, nm_id, brand, status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id`,
			item.ChrtID, item.TrackNumber, item.Price, item.Rid, item.Name, item.Sale, item.Size, item.TotalPrice,
			item.NmID, item.Brand, item.Status).Scan(&lastInsertId)
		if err != nil {
			log.Printf("%v: can`t insert item to items table: %v\n", db.name, err)
			return -1, err
		}
		itemsIds = append(itemsIds, lastInsertId)
	}

	err = tx.QueryRow(context.Background(), `INSERT INTO payments (transaction, request_id, currency, provider, amount,
		payment_dt, bank, delivery_cost, goods_total, custom_fee) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`,
		o.Payment.Transaction, o.Payment.RequestID, o.Payment.Currency, o.Payment.Provider, o.Payment.Amount,
		o.Payment.PaymentDt, o.Payment.Bank, o.Payment.DeliveryCost, o.Payment.GoodsTotal, o.Payment.CustomFee).Scan(&lastInsertId)
	if err != nil {
		log.Printf("%v: can`t insert payment to payments table: %v\n", db.name, err)
		return -1, err
	}
	paymentIdFk := lastInsertId

	err = tx.QueryRow(context.Background(), `INSERT INTO deliverys (name, phone, zip, city, address, region, email)
		values ($1, $2, $3, $4, $5, $6, $7) RETURNING id`, o.Delivery.Name, o.Delivery.Phone, o.Delivery.Zip,
		o.Delivery.City, o.Delivery.Address, o.Delivery.Region, o.Delivery.Email).Scan(&lastInsertId)
	if err != nil {
		log.Printf("%v: can`t insert delivery to deliverys table: %v\n", db.name, err)
		return -1, err
	}
	deliveryIdFk := lastInsertId

	err = tx.QueryRow(context.Background(), `INSERT INTO orders (order_uid, track_number, entry, delivery_id_fk, 
        payment_id_fk, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, oof_shard) 
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING id`, o.OrderUID, o.TrackNumber, o.Entry,
		deliveryIdFk, paymentIdFk, o.Locale, o.InternalSignature, o.CustomerID, o.DeliveryService, o.ShardKey, o.SmID,
		o.OofShard).Scan(&lastInsertId)
	if err != nil {
		log.Printf("%v: can`t insert order to orders table: %v\n", db.name, err)
		return -1, err
	}
	orderIdFk := lastInsertId

	for _, itemId := range itemsIds {
		_, err := tx.Exec(context.Background(), `INSERT INTO order_items (order_id_fk, item_id_fk) values ($1, $2)`,
			orderIdFk, itemId)
		if err != nil {
			log.Printf("%v: : can`t insert order_id_fk, item_id_fk to order_items table: %v\n", db.name, err)
			return -1, err
		}
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return 0, err
	}

	log.Printf("%v: add order `%v` to DB - OK\n", db.name, o.OrderUID)
	db.csh.SetOrder(orderIdFk, o)
	return orderIdFk, nil
}

func (db *DB) PutOrderIDToCache(oid int64) {
	db.pool.QueryRow(context.Background(), `INSERT INTO cache (order_id, cache_stream) VALUES ($1, $2)`, oid, os.Getenv("DB_CACHE_STREAM"))
	log.Printf("%v: put order ID `%v` to cache table - OK\n", db.name, oid)
}

func (db *DB) DeleteCache() {
	_, err := db.pool.Exec(context.Background(), `DELETE FROM cache WHERE cache_stream = $1`, os.Getenv("DB_CACHE_STREAM"))
	if err != nil {
		log.Printf("%v: delete cache stream '%v' from DB - OK\n", db.name, os.Getenv("DB_CACHE_STREAM"))
	}
	log.Printf("%v: clear cache stream '%v' from DB - OK\n", db.name, os.Getenv("DB_CACHE_STREAM"))
}
