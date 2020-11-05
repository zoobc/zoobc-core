package storage

type (
	CacheStorageInterface interface {
		// SetItem take any item and store to its specific storage implementation
		SetItem(key, item interface{}) error
		// GetItem take variable and assign implementation stored item to it
		GetItem(key, item interface{}) error
		// GetAllItems fetch all cached items
		GetAllItems(item interface{}) error
		// RemoveItem remove item by providing the key(s)
		RemoveItem(key interface{}) error
		// GetSize return the size of storage in number of `byte`
		GetSize() int64
		// ClearCache empty the storage item
		ClearCache() error
	}
)
