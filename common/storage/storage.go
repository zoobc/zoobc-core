package storage

type (
	CacheStorageInterface interface {
		// SetItem take any item and store to its specific storage implementation
		SetItem(item interface{}) error
		// GetItem take variable and assign implementation stored item to it
		GetItem(interface{}) error
		// GetSize return the size of storage in number of `byte`
		GetSize() int64
		// ClearCache empty the storage item
		ClearCache() error
	}
)
