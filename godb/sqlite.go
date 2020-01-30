package db

import (
	"fmt"
	"sync"

	"github.com/jinzhu/gorm"
)

type SqliteMaster struct {
	Type     string `json:"type"`
	TblName  string `json:"tbl_name"`
	RootPage string `json:"root_page"`
	Name     string `json:"name"`
	Sql      string `json:"sql"`
}
type TableInfo struct {
	Cid       uint        `json:"cid"`
	Type      string      `json:"type"`
	Name      string      `json:"name"`
	Notnull   string      `json:"not_null"`
	Pk        int         `json:"pk"`
	DfltValue interface{} `json:"deflt_value"`
}
type Sqlite struct {
	orm   *gorm.DB
	mutex *sync.Mutex
}

func (Sqlite) Hello() string {
	return "Sqlite"
}
func (s Sqlite) Orm() *gorm.DB {
	return s.orm
}
func (s Sqlite) Mutex() *sync.Mutex {
	return s.mutex
}
func (db Sqlite) Insert(obj interface{}) error {
	tx := db.orm.Begin()

	err := tx.Create(obj).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return err
	}
	return nil
}
func (db Sqlite) Delete(obj interface{}, para map[string]interface{}) error {
	tx := db.orm.Begin()

	err := tx.Where(para).Delete(obj).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return err
	}
	return nil
}
func (db Sqlite) Get(objs interface{}, where map[string]interface{}, order string, limit uint64, offset uint64, deleted bool) error {
	con := db.orm
	if deleted {
		con = con.Unscoped()
	}

	keys, values := whereExactKeyValues(where)
	con = con.Where(keys, values)
	if deleted {
		con = con.Where("deleted_at is not null")
	}

	if order != "" {
		con = con.Order(order)
	}
	if limit != 0 {
		con = con.Limit(limit)
	}
	if offset != 0 {
		con = con.Offset(offset)
	}
	err := con.Find(objs).Error
	if err != nil {
		return err
	}
	return nil
}
func (db Sqlite) Update(obj interface{}, para map[string]interface{}) error {
	tx := db.orm.Begin()

	err := tx.Model(obj).Updates(para).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return err
	}
	return nil
}
func (db Sqlite) GetTables() (interface{}, error) {
	var sqlMaster []SqliteMaster
	if err := db.Get(&sqlMaster, map[string]interface{}{"type": "table"}, "", 0, 0, false); err != nil {
		return nil, err
	}
	return sqlMaster, nil
}
func (db Sqlite) GetMeta(tableName interface{}) (interface{}, error) {
	var info []TableInfo
	str := fmt.Sprint(tableName)
	q := "pragma table_info(" + str + ")"
	row, err := db.orm.DB().Query(q)
	if err != nil {
		return info, err
	}
	var cid uint
	var t string
	var name string
	var notnull string
	var pk int
	var dfltValue interface{}
	for row.Next() {
		err = row.Scan(&cid, &name, &t, &notnull, &dfltValue, &pk)
		if err != nil {
			return info, err
		}
		table_info := TableInfo{
			Cid:       cid,
			Name:      name,
			Type:      t,
			Notnull:   notnull,
			DfltValue: dfltValue,
			Pk:        pk,
		}
		info = append(info, table_info)
	}
	return info, nil
}
