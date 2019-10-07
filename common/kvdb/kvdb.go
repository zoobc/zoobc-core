/**
kvdb is key-value database abstraction of badger db implementation
*/
package kvdb

import (
	"time"

	"github.com/dgraph-io/badger"
)

type (
	KVExecutorInterface interface {
		Insert(key string, value []byte, expiry int) error
		BatchInsert(updates map[string][]byte) error
		Get(key string) ([]byte, error)
		GetByPrefix(prefix string) (map[string][]byte, error)
	}
	KVExecutor struct {
		Db *badger.DB
	}
)

func NewKVExecutor(db *badger.DB) *KVExecutor {
	return &KVExecutor{
		Db: db,
	}
}

// Insert insert a single record of data by providing the key in string and value in []byte
// expiry represent number of minutes
func (kve *KVExecutor) Insert(key string, value []byte, expiry int) error {
	err := kve.Db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte(key), value).WithTTL(time.Duration(expiry) * time.Minute)
		err := txn.SetEntry(e)
		return err
	})
	if err != nil {
		return err
	}
	return nil
}

// BatchInsert insert multiple record of data accepting map of string - []byte
func (kve *KVExecutor) BatchInsert(updates map[string][]byte) error {
	err := kve.Db.Update(func(txn *badger.Txn) error {
		for k, v := range updates {
			err := txn.Set([]byte(k), v)
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// Get accept string key and do a lookup on the kvdb for an item
func (kve *KVExecutor) Get(key string) ([]byte, error) {
	var valCopy []byte
	err := kve.Db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		valCopy, err = item.ValueCopy(valCopy)
		return err
	})
	if err != nil {
		return nil, err
	}
	return valCopy, nil
}

// GetByPrefix will search through our kvdb for key that're prefixed with `prefix` and return the
// unsorted results
func (kve *KVExecutor) GetByPrefix(prefix string) (map[string][]byte, error) {
	var result = make(map[string][]byte)
	err := kve.Db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek([]byte(prefix)); it.ValidForPrefix([]byte(prefix)); it.Next() {
			item := it.Item()
			k := item.Key()
			err := item.Value(func(v []byte) error {
				result[string(k)] = v
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
