package db

import (
	"fmt"
	"sync"
)

type DB interface {
	WarnAdd(chat_id int64, user_id int) error
	WarnSet(chat_id int64, user_id, count int) error
	WarnCount(chat_id int64, user_id int) (int, error)
	WarnMax(chat_id int64) (int, error)
	WarnMaxSet(chat_id int64, count int) error
	WordLogWlGet() ([]string, error)
	WordLogWlAdd(word string) error
	WordLogWlDel(word string) error
	WordLogBlGet() ([]string, error)
	WordLogBlAdd(word string) error
	WordLogBlDel(word string) error
	RulesGet(chat_id int64) (string, error)
	RulesSet(chat_id int64, rules string) error
}
type dbFactory interface {
	New(string) (DB, error)
}

var (
	factoryMu sync.RWMutex
	factories = make(map[string]dbFactory)
)

func Register(name string, factory dbFactory) {
	factoryMu.Lock()
	defer factoryMu.Unlock()
	if _, dup := factories[name]; dup {
		panic("db: Register called twice for driver " + name)
	}
	factories[name] = factory
}

func New(driver string, source string) (DB, error) {
	factoryMu.RLock()
	defer factoryMu.RUnlock()
	factory, ok := factories[driver]
	if !ok {
		return nil, fmt.Errorf("db: Unknown driver '%s'", driver)
	}

	return factory.New(source)
}
