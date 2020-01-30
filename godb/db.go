package db

import (
	"errors"
	"sync"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

const (
	SQLITE_DRIVER   = "sqlite3"
	POSTGRES_DRIVER = "postgres"
)

type DB interface {
	Hello() string
	Orm() *gorm.DB
	Mutex() *sync.Mutex
	Insert(obj interface{}) error
	Get(objs interface{}, where map[string]interface{}, order string, limit uint64, offset uint64, deleted bool) error
	Update(obj interface{}, para map[string]interface{}) error
	Delete(obj interface{}, para map[string]interface{}) error
	GetMeta(name interface{}) (interface{}, error)
	GetTables() (interface{}, error)
}

//support databes 'sqlite3','postgres'
func Open(driver string, args ...interface{}) (DB, error) {
	if driver == "" {
		return nil, errors.New("Invalid driver name.")
	}
	d, err := gorm.Open(driver, args...)
	if err != nil {
		return nil, err
	}
	var db DB
	switch driver {
	case SQLITE_DRIVER:
		db = Sqlite{orm: d, mutex: &sync.Mutex{}}
	case POSTGRES_DRIVER:
		db = Postgres{orm: d, mutex: &sync.Mutex{}}
	default:
		return nil, errors.New("Unknown database driver")
	}
	return db, nil
}
