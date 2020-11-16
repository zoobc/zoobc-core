package storage

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"sync"

	"github.com/zoobc/zoobc-core/common/blocker"
)

type (
	// SpendableBalanceStorage cache for spendable balance
	SpendableBalanceStorage struct {
		sync.RWMutex
		spendableBalances map[string]int64
	}

	SpendableBalaceCacheObject struct {
		AccountAddress   []byte
		SpendableBalance int64
	}
)

func NewSpendableBalanceStorage() *SpendableBalanceStorage {
	return &SpendableBalanceStorage{
		spendableBalances: make(map[string]int64),
	}
}

func (sp *SpendableBalanceStorage) SetItem(key, item interface{}) error {
	account, ok := key.([]byte)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongTypeKeyExpected:[]byte")
	}
	spendableBalace, ok := item.(int64)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongTypeKeyExpected:int64")
	}
	sp.Lock()
	sp.spendableBalances[fmt.Sprintf("%q", account)] = spendableBalace
	sp.Unlock()
	return nil
}

func (sp *SpendableBalanceStorage) SetItems(_ interface{}) error {
	return nil
}

func (sp *SpendableBalanceStorage) GetItem(key, item interface{}) error {
	account, ok := key.([]byte)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongTypeKeyExpected:[]byte")
	}
	spendableBalace, ok := item.(*int64)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongTypeKeyExpected:*int64")
	}
	sp.RLock()
	defer sp.RUnlock()
	*spendableBalace = sp.spendableBalances[fmt.Sprintf("%q", account)]
	if *spendableBalace == 0 {
		return blocker.NewBlocker(blocker.NotFound, "SpendableBalanceStorageZero")
	}
	return nil
}

func (sp *SpendableBalanceStorage) GetAllItems(item interface{}) error {
	spendableBalacesObject, ok := item.(*[]SpendableBalaceCacheObject)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongTypeKeyExpected:*[]SpendableBalaceCacheObject")
	}
	for accountStr, spendable := range sp.spendableBalances {
		*spendableBalacesObject = append(*spendableBalacesObject, SpendableBalaceCacheObject{
			AccountAddress:   []byte(accountStr),
			SpendableBalance: spendable,
		})
	}
	return nil
}
func (sp *SpendableBalanceStorage) GetTotalItems() int {
	return len(sp.spendableBalances)
}

func (sp *SpendableBalanceStorage) RemoveItem(key interface{}) error {
	AccountPubKey, ok := key.([]byte)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongTypeKeyExpected:[]byte")
	}
	sp.Lock()
	delete(sp.spendableBalances, fmt.Sprintf("%q", AccountPubKey))
	sp.Unlock()
	return nil
}

func (sp *SpendableBalanceStorage) size() int {
	var (
		nBytes bytes.Buffer
		enc    = gob.NewEncoder(&nBytes)
	)
	_ = enc.Encode(sp.spendableBalances)
	return nBytes.Len()
}

func (sp *SpendableBalanceStorage) GetSize() int64 {
	return int64(sp.size())
}
func (sp *SpendableBalanceStorage) ClearCache() error {
	sp.spendableBalances = make(map[string]int64)
	return nil
}
