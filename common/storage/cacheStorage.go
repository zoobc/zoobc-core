package storage

type (
	CacheStorageInterface interface {
		SetItem(item interface{}) error
		GetItem(interface{}) error
		GetSize() int64
		ClearCache() error
	}
)
