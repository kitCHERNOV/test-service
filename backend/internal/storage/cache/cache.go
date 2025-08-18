package cache

import (
	"test/internal/config"
	"test/internal/models"
)

type Cache map[string]models.Order

// кэш для хранения
func InitCache(params config.CacheParams) Cache {
	return make(Cache, params.Amount)
}

func (c Cache) CacheItemAdd(key string, item models.Order) {
	c[key] = item
}
