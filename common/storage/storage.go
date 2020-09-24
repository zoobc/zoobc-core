package storage

type (
	CacheStorageInterface interface {
		// SetItem take any item and store to its specific storage implementation
		SetItem(key, item interface{}) error
		// SetItems take all items that stored and refill item reference
		SetItems(item interface{}) error
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

	CacheStackStorageInterface interface {
		// Pop delete the latest item on the stack
		Pop() error
		// Push item into the stack, if exceed size first item is deleted and shifted
		Push(interface{}) error
		// PopTo takes index (uint32) and delete item to the given index (start from 0)
		PopTo(uint32) error
		// GetAll return all item in the stack to given `interface` arguments
		GetAll(interface{}) error
		// GetAtIndex return item at given index
		GetAtIndex(uint32, interface{}) error
		// GetTop return top item on the stack
		GetTop(interface{}) error
		// Clear clean up the whole stack and reinitialize with new array
		Clear() error
	}

	TransactionalCache interface {
		// Begin prepare state of cache for transactional writes, must called at start of tx writes
		Begin() error
		// Commit finalize transactional writes to the struct
		Commit() error
		// Rollback release locks and return state of struct to original before tx modifications are made
		Rollback() error
		// TxSetItem set individual item
		TxSetItem(id, item interface{}) error
		// TxSetItems replace items in bulk
		TxSetItems(items interface{}) error
		// TxRemoveItem remove item with given ID
		TxRemoveItem(id interface{}) error
	}
)
