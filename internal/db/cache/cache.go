package cache

import (
	"log"
	"os"
	"prodWBRL0/internal/db"
	"prodWBRL0/internal/db/models"
	"strconv"
	"sync"
)

type Cache struct {
	buffer   map[int64]models.Order
	priority []int64
	bufSize  int
	ind      int
	DBInst   *db.DB
	name     string
	mutex    *sync.RWMutex
}

func NewCache(db *db.DB) *Cache {
	csh := Cache{}
	csh.Init(db)
	return &csh
}

func (c *Cache) Init(db *db.DB) {
	c.DBInst = db
	db.SetCacheInstance(c)
	c.name = "Cache"
	c.mutex = &sync.RWMutex{}
	bufSize, err := strconv.Atoi(os.Getenv("DB_CACHE_STREAM_SIZE"))
	if err != nil {
		log.Printf("%s: set DB_CACHE_STREAM_SIZE to 5 orders \n", c.name)
		bufSize = 5
	}
	c.bufSize = bufSize
	c.buffer = make(map[int64]models.Order, c.bufSize)
	c.priority = make([]int64, c.bufSize)
	c.getCacheFromDatabase()
}

func (c *Cache) getCacheFromDatabase() {
	log.Printf("%v: get cache from DB...\n", c.name)
	buf, queue, pos, err := c.DBInst.GetCacheState(c.bufSize)
	if err != nil {
		log.Printf("%s: can't get cache from DB: %v\n", c.name, err)
		return
	}

	if pos == c.bufSize {
		pos = 0
	}
	c.mutex.Lock()
	c.buffer = buf
	c.priority = queue
	c.ind = pos
	c.mutex.Unlock()
	log.Printf("%s: get cache from DB, priority is: `%v`, next priority `%v`", c.name, c.priority, c.ind)
}

func (c *Cache) SetOrder(oid int64, o models.Order) {
	if c.bufSize > 0 {
		c.mutex.Lock()
		c.priority[c.ind] = oid
		c.ind++
		if c.ind == c.bufSize {
			c.ind = 0
		}
		c.buffer[oid] = o
		c.mutex.Unlock()
		c.DBInst.PutOrderIDToCache(oid)
		log.Printf("%s: add order ID '%v' to cache table - OK, queue is %v\n", c.name, oid, c.ind)
	} else {
		log.Printf("%s: can`t init cache, DB_CACHE_STREAM_SIZE = 0 \n", c.name)
	}
	log.Printf("%s: queue is: %v, next position in queue is: %v", c.name, c.priority, c.ind)
}

func (c *Cache) DeleteCache() {
	log.Printf("%s: start DeleteCache()...", c.name)
	c.DBInst.DeleteCache()
	log.Printf("%s: DeleteCache() - OK", c.name)
}
