package cache

import (
	"encoding/json"
	"log"
	"os"
	"test/internal/config"
	"test/internal/models"
	"test/internal/storage/postgres"
)

type Cache map[string]models.Order

type FifoCache struct {
	capacity int
	data     Cache
	ring     []string
	pos      int
	size     int
	alive    map[string]bool
}

// TODO: Переписать функцию так чтобы был вариант востановления кжша
func NewFifoCache(capacity int) *FifoCache {
	return &FifoCache{
		capacity: capacity,
		data:     make(Cache, capacity),
		ring:     make([]string, capacity),
		alive:    make(map[string]bool, capacity),
	}
}

// Востановление кэша
func RestoreCache(path string, capacity int, storage *postgres.Storage) *FifoCache {
	const op = "storage.cache.RestoreCache"
	uids, err := LoadCacheMetaData(path)
	if err != nil {
		log.Println("Loc: %s, Err: %v", op, err)
		return nil
	}
	orders, err := storage.GetDataToRestoreCache(uids)
	if err != nil {
		log.Println("Loc: %s, Err: %v", op, err)
		return nil
	}
	cache := &FifoCache{
		capacity: capacity,
		data:     orders,
		ring:     make([]string, capacity),
		alive:    make(map[string]bool, capacity),
	}
	i := 0
	for uid := range orders {
		cache.ring[i] = uid
		cache.alive[uid] = true
		i++
	}
	log.Println("Cache is loaded")
	return cache
}

func (c *FifoCache) Set(cfg *config.Config, key string, val models.Order) {
	if _, ok := c.data[key]; ok {
		c.data[key] = val

		err := SaveCacheMetaData(cfg.CacheParams.Path, c.data)
		if err != nil {
			log.Println("Troubles with saving restore cache metadata")
		}
		return
	}
	if c.size < c.capacity {
		c.data[key] = val
		c.ring[c.pos] = key
		c.alive[key] = true
		c.pos = (c.pos + 1) % c.capacity
		c.size++

		// Снова попытка записи
		err := SaveCacheMetaData(cfg.CacheParams.Path, c.data)
		if err != nil {
			log.Println("Troubles with saving restore cache metadata")
		}
		return
	}
	oldKey := c.ring[c.pos]
	if c.alive[oldKey] {
		delete(c.data, oldKey)
		delete(c.alive, oldKey)
	}
	c.data[key] = val
	c.ring[c.pos] = key
	c.alive[key] = true
	c.pos = (c.pos + 1) % c.capacity

	// Сохранение в json нового набора ключей
	err := SaveCacheMetaData(cfg.CacheParams.Path, c.data)
	if err != nil {
		log.Println("Troubles with saving restore cache metadata")
	}
}

func (c *FifoCache) Get(key string) (models.Order, bool) {
	v, ok := c.data[key]
	log.Println("Кэш отдал данные")
	return v, ok
}

func (c *FifoCache) Delete(key string) bool {
	if _, ok := c.data[key]; !ok {
		return false
	}
	delete(c.data, key)
	if c.alive[key] {
		delete(c.alive, key)
		c.size--
	}
	return true
}

// Функции сохранения метаданных и востановления

func SaveCacheMetaData(path string, data Cache) error {
	var uids = []string{}
	// Сохраним все доступные uid параметры
	for uid := range data {
		uids = append(uids, uid)
	}

	bytes, err := json.Marshal(uids)
	if err != nil {
		return err
	}
	return os.WriteFile(path, bytes, 0644)
}

// Востановление кэша из бд по сохроненным метаданным
func LoadCacheMetaData(path string) ([]string, error) {
	var titles []string
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, &titles)
	if err != nil {
		return nil, err
	}
	return titles, nil
}
