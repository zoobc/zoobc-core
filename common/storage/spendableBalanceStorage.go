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
		transactionalLock              sync.RWMutex
		isInTransaction                bool
		transactionalSpendableBalances map[string]int64
		spendableBalances              map[string]int64
	}

	spendableBalanceCacheObject struct {
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
	spendableBalance, ok := item.(int64)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongTypeKeyExpected:int64")
	}
	sp.Lock()
	sp.spendableBalances[fmt.Sprintf("%q", account)] = spendableBalance
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
	spendableBalance, ok := item.(*int64)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongTypeKeyExpected:*int64")
	}

	if sp.isInTransaction {
		// return from transactional list
		sp.transactionalLock.RLock()
		*spendableBalance = sp.transactionalSpendableBalances[fmt.Sprintf("%q", account)]
		sp.transactionalLock.RUnlock()
	} else {
		// return from normal list
		sp.RLock()
		*spendableBalance = sp.spendableBalances[fmt.Sprintf("%q", account)]
		sp.RUnlock()
	}
	return nil
}

func (sp *SpendableBalanceStorage) GetAllItems(item interface{}) error {
	spendableBalancesObject, ok := item.(*[]spendableBalanceCacheObject)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongTypeKeyExpected:*[]spendableBalanceCacheObject")
	}
	if sp.isInTransaction {
		sp.transactionalLock.RLock()
		for accountStr, spendable := range sp.transactionalSpendableBalances {
			*spendableBalancesObject = append(*spendableBalancesObject, spendableBalanceCacheObject{
				AccountAddress:   []byte(accountStr),
				SpendableBalance: spendable,
			})
		}
		sp.transactionalLock.RUnlock()
	} else {
		sp.RLock()
		for accountStr, spendable := range sp.spendableBalances {
			*spendableBalancesObject = append(*spendableBalancesObject, spendableBalanceCacheObject{
				AccountAddress:   []byte(accountStr),
				SpendableBalance: spendable,
			})
		}
		sp.RUnlock()
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

// Transactional implementation

// Begin prepare data to begin doing transactional change to the cache, this implementation
// will never return error
func (sp *SpendableBalanceStorage) Begin() error {
	sp.Lock()
	sp.transactionalLock.Lock()
	defer sp.transactionalLock.Unlock()
	sp.isInTransaction = true
	sp.transactionalSpendableBalances = make(map[string]int64)
	// copy current spendable balance into transaction list
	for accountStr, spendableBalance := range sp.spendableBalances {
		sp.transactionalSpendableBalances[accountStr] = spendableBalance
	}
	return nil
}

func (sp *SpendableBalanceStorage) Commit() error {
	// make sure isInTransaction is true
	if !sp.isInTransaction {
		return blocker.NewBlocker(blocker.ValidationErr, "BeginIsRequired")
	}
	sp.transactionalLock.Lock()
	defer func() {
		sp.isInTransaction = false
		sp.Unlock()
		sp.transactionalLock.Unlock()
	}()
	// Update all spendable belence from transactional spendable belance
	for accountStr, spendableBalance := range sp.transactionalSpendableBalances {
		sp.spendableBalances[accountStr] = spendableBalance
	}
	sp.transactionalSpendableBalances = make(map[string]int64)
	return nil
}

func (sp *SpendableBalanceStorage) Rollback() error {
	// make sure isInTransaction is true
	if !sp.isInTransaction {
		return blocker.NewBlocker(blocker.ValidationErr, "BeginIsRequired")
	}
	sp.transactionalLock.Lock()
	defer func() {
		sp.isInTransaction = false
		sp.Unlock()
		sp.transactionalLock.Unlock()
	}()
	sp.transactionalSpendableBalances = make(map[string]int64)
	return nil
}

// TxSetItem set individual item
func (sp *SpendableBalanceStorage) TxSetItem(id, item interface{}) error {
	// make sure isInTransaction is true
	if !sp.isInTransaction {
		return blocker.NewBlocker(blocker.ValidationErr, "BeginIsRequired")
	}
	account, ok := id.([]byte)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongTypeKeyExpected:[]byte")
	}
	spendableBalance, ok := item.(int64)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongTypeKeyExpected:int64")
	}
	sp.transactionalLock.Lock()
	sp.transactionalSpendableBalances[fmt.Sprintf("%q", account)] = spendableBalance
	sp.transactionalLock.Unlock()
	return nil
}

// TxSetItems currently doesn’t need to set in transactional
func (sp *SpendableBalanceStorage) TxSetItems(items interface{}) error {
	return blocker.NewBlocker(blocker.ValidationErr, "NotYetImeplemented")
}

// TxRemoveItem currently doesn’t need to remove in transactional
func (sp *SpendableBalanceStorage) TxRemoveItem(id interface{}) error {
	return blocker.NewBlocker(blocker.ValidationErr, "NotYetImeplemented")
}
