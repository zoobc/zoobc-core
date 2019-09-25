/**
kvdb is key-value database abstraction of badger db implementation
*/
package kvdb

type (
	KVExecutorInterface interface {
	}
	KVExecutor struct {
	}
)

func (*KVExecutor) Insert(key string, value interface{}) error {

	return nil
}

func (*KVExecutor) Get(key string) (interface{}, error) {
	return nil, nil
}
